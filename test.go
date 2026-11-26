package main

import (
	"context"

	"dagger/pomerium/internal/dagger"
)

// Test runs the Go tests (with the race detector) for the repository's main
// module.
//
// +check
func (m *Pomerium) Test(ctx context.Context) error {
	coverage, err := m.Cover(ctx)
	if err != nil {
		return err
	}
	_, err = coverage.Sync(ctx)
	return err
}

// Cover runs the Go tests like Test and returns the coverage profile.
func (m *Pomerium) Cover(ctx context.Context) (*dagger.File, error) {
	ctr, err := m.GoBase(ctx)
	if err != nil {
		return nil, err
	}
	return ctr.
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@" + gotestsumVersion}).
		WithServiceBinding("docker", dockerService()).
		WithEnvVariable("DOCKER_HOST", "tcp://docker:2375").
		// Ryuk cleans up containers when tests exit; the Docker service is
		// thrown away instead.
		WithEnvVariable("TESTCONTAINERS_RYUK_DISABLED", "true").
		WithExec([]string{"sh", "-euc", `
# Dagger sets OpenTelemetry variables to collect traces from commands; tests
# that read the OpenTelemetry configuration from the environment expect none.
for v in $(env | sed -n 's/^\(OTEL_[A-Z_]*\)=.*/\1/p') TRACEPARENT; do
	unset "$v"
done

mkdir -p /tmp/coverdata
gotestsum --rerun-fails --packages=./... -- \
	-p "$TEST_PARALLELISM" -race -cover -args -test.gocoverdir=/tmp/coverdata
go tool covdata textfmt -i=/tmp/coverdata -o /tmp/coverage.txt
sed -i \
	-e '/\.pb\.go\:/d' \
	-e '/\/mock_.*\.go\:/d' \
	-e '/\.gen\.go\:/d' \
	-e '/\/internal\/testenv\/:/d' \
	-e '/\/internal\/testutil\/:/d' \
	/tmp/coverage.txt
`}).
		File("/tmp/coverage.txt"), nil
}

// dockerService returns a Docker daemon for tests that start containers.
func dockerService() *dagger.Service {
	return dag.Container().
		From("docker:"+dockerVersion+"-dind").
		WithEnvVariable("DOCKER_TLS_CERTDIR", "").
		// Docker's storage can't be on the container's own (overlay) filesystem.
		// A cache volume also keeps pulled images between runs; Private gives
		// concurrent runs their own copy, since daemons can't share one.
		WithMountedCache("/var/lib/docker", dag.CacheVolume("docker-dind"), dagger.ContainerWithMountedCacheOpts{
			Sharing: dagger.CacheSharingModePrivate,
		}).
		WithExposedPort(2375).
		AsService(dagger.ContainerAsServiceOpts{
			Args:                     []string{"dockerd", "--host=tcp://0.0.0.0:2375", "--tls=false"},
			UseEntrypoint:            true,
			InsecureRootCapabilities: true,
		})
}
