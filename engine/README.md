# Self-hosted Dagger engine for CI

GitHub-hosted runners start every job with an empty Dagger engine, so nothing is
cached between CI runs. This directory sets up a single long-lived engine on
Kubernetes, with a persistent volume for its cache, that CI connects to over
mutual TLS.

```
GitHub-hosted runner ── tls://HOST:1234 (client cert) ──▶ LoadBalancer ──▶ engine pod (StatefulSet + PVC)
```

| File | Purpose |
| --- | --- |
| `values.yaml` | Values for the official Dagger Helm chart |
| `gen-certs.sh` | Generates the CA, server and client certificates |
| `networkpolicy.yaml` | Example NetworkPolicy restricting what CI jobs can reach |

Why not run the CI runners in the cluster (Actions Runner Controller) and share
an engine per node, as Dagger recommends? Pomerium's repositories are public, and
self-hosted runners would run code from fork pull requests on our
infrastructure.

## Setup

### 1. Nodes

Run the engine on dedicated nodes with fast local disks. Label and taint a node
pool, then uncomment `nodeSelector` and `tolerations` in `values.yaml` to match.

### 2. Certificates

```sh
./gen-certs.sh dagger-engine.example.com ./certs
kubectl create namespace dagger
kubectl -n dagger create secret generic dagger-engine-tls \
  --from-file=./certs/tls.crt --from-file=./certs/tls.key --from-file=./certs/ca.crt
```

Pass the hostname CI will connect to; the server certificate is only valid for
that name. Keep `ca.key` offline, since it is only needed to issue new
certificates. Never commit `certs/`.

### 3. Engine

```sh
helm upgrade --install --namespace dagger checks oci://registry.dagger.io/dagger-helm \
  --version 1.0.0-beta.30 -f values.yaml
kubectl -n dagger rollout status statefulset/checks-dagger-helm-engine
```

This creates:

- a one-replica StatefulSet with a 200Gi PersistentVolumeClaim for the cache;
- a `LoadBalancer` Service named `checks-dagger-helm` on port 1234.

The engine prunes its cache at 75% of the disk by default. To change that, set
`engine.configJson`; see Dagger's `engine.json` reference.

Point the hostname's DNS at the LoadBalancer address:

```sh
kubectl -n dagger get service checks-dagger-helm
```

### 4. Network policy

Edit the placeholder CIDRs in `networkpolicy.yaml` to your cluster's pod and
service ranges, then apply it:

```sh
kubectl apply -f networkpolicy.yaml
```

### 5. Verify

From a machine with the client certificate:

```sh
DAGGER_ENGINE="tls://dagger-engine.example.com:1234?ca_path=certs/ca.crt&cert_path=certs/client.crt&key_path=certs/client.key" \
  dagger check -l
```

A client without a valid certificate does not get an error: the CLI keeps
retrying, and the engine logs `tls: client didn't provide a certificate`.

### 6. GitHub configuration

In each repository (or once for the organization), add the secrets
`DAGGER_CA_CERT` (`ca.crt`), `DAGGER_CLIENT_CERT` (`client.crt`) and
`DAGGER_CLIENT_KEY` (`client.key`).

Workflows run Dagger through this repository's action (`../actions/run/action.yml`), which
connects to the shared engine and installs the matching Dagger version. Pin it
to a commit, and set `timeout-minutes` on the job, because a rejected
certificate hangs instead of failing:

```yaml
jobs:
  govulncheck:
    runs-on: ubuntu-latest
    timeout-minutes: 20
    steps:
      - uses: actions/checkout@v4
      - uses: calebdoxsey/pomerium-dagger-poc/actions/run@<commit>
        with:
          args: --check govulncheck
          engine-ca-cert: ${{ secrets.DAGGER_CA_CERT }}
          engine-client-cert: ${{ secrets.DAGGER_CLIENT_CERT }}
          engine-client-key: ${{ secrets.DAGGER_CLIENT_KEY }}
```

Fork pull requests do not receive secrets. The action then uses a temporary
local engine, as it would without this setup.

## Upgrading Dagger

The chart, the engine and the CLI must be the same version: `daggerVersion` in
`../version.go`. Every check fails if the engine it runs on is a different
version. To upgrade:

1. Change `daggerVersion` and the `version` in `../actions/run/action.yml`
   (`go test ./...` fails if they differ).
2. Upgrade the Helm release (`--version`) or engine deployment.
3. Update each repository to the new commit of this action and module
   (`dagger lock update`).

## Security

- **The engine is privileged.** Anyone with a client certificate can run
  arbitrary privileged containers on its node. Keep it on dedicated nodes and
  apply the NetworkPolicy, so CI jobs cannot reach in-cluster services or the
  cloud metadata endpoint.
- **Repository write access means engine access.** Anyone who can push a branch
  can run a workflow that reads the client certificate. GitHub environments with
  protection rules can narrow this, but they also block pull request branches.
- **Shared cache volumes can be poisoned.** The Go module and build caches are
  mutable and shared by every job. A malicious pull request could plant
  build-cache entries that a later `main` or release build uses. Before release
  builds use this setup, install a second Helm release (for example `trusted`)
  used only by `main` and tags, so trusted and untrusted work never share a
  cache.

## Scaling

Start with one engine and give it a larger node when it gets slow. Each engine
has its own cache. To add engines, either give a check its own Helm release
(and runner configuration), or put several engines behind a Service with
`sessionAffinity: ClientIP`, because one CLI session opens several connections
that must reach the same engine.
