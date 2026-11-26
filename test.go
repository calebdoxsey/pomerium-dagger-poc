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
		With(func(c *dagger.Container) *dagger.Container {
			return withScript(c, "cover.sh", "/tmp/coverage.txt")
		}).
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
			// dockerd logs every container start and stop at info level, which
			// floods the CI log; warn keeps real problems. (Its startup warnings
			// about the unencrypted API are expected: only the tests can reach it.)
			Args:                     []string{"dockerd", "--host=tcp://0.0.0.0:2375", "--tls=false", "--log-level=warn"},
			UseEntrypoint:            true,
			InsecureRootCapabilities: true,
		})
}
