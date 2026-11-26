package main

import (
	"context"

	"dagger/pomerium/internal/dagger"
)

// GoBase returns a container with the Go toolchain installed, the prepared
// source mounted at /src and the Go module and build caches attached.
//
// The Go version is goVersion. Go won't download a different toolchain, so a
// repository whose go.mod requires a newer Go fails instead.
func (m *Pomerium) GoBase(ctx context.Context) (*dagger.Container, error) {
	source, err := m.Prepare(ctx)
	if err != nil {
		return nil, err
	}
	return m.goContainer(ctx, source)
}

func (m *Pomerium) goContainer(ctx context.Context, source *dagger.Directory) (*dagger.Container, error) {
	if err := m.requireEngineVersion(ctx); err != nil {
		return nil, err
	}
	return dag.Container().
		From("golang:"+goVersion).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithEnvVariable("GOCACHE", "/root/.cache/go-build").
		WithEnvVariable("GOTOOLCHAIN", "local").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src"), nil
}
