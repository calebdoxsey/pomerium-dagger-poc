package main

import (
	"context"

	"dagger/pomerium/internal/dagger"
)

// Prepare returns the source with any build dependencies that are not checked
// into git fetched (the equivalent of `make deps-build`).
//
// Currently this downloads the embedded envoy binaries for repositories that
// contain pkg/envoy/get-envoy.
func (m *Pomerium) Prepare(ctx context.Context) (*dagger.Directory, error) {
	hasGetEnvoy, err := m.Source.Exists(ctx, "pkg/envoy/get-envoy")
	if err != nil {
		return nil, err
	}
	if !hasGetEnvoy {
		return m.Source, nil
	}

	ctr, err := m.goContainer(ctx, m.Source)
	if err != nil {
		return nil, err
	}
	envoyFiles := ctr.
		WithWorkdir("/src/pkg/envoy/files").
		WithExec([]string{"go", "run", "../get-envoy"}).
		Directory("/src/pkg/envoy/files")
	return m.Source.WithDirectory("pkg/envoy/files", envoyFiles), nil
}
