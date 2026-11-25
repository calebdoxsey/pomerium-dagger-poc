package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"dagger/pomerium/internal/dagger"
)

// GoBase returns a container with the Go toolchain installed, the prepared
// source mounted at /src and the Go module and build caches attached.
//
// The Go version is read from the "golang" entry in the repository's
// .tool-versions file.
func (m *Pomerium) GoBase(ctx context.Context) (*dagger.Container, error) {
	source, err := m.Prepare(ctx)
	if err != nil {
		return nil, err
	}
	return m.goContainer(ctx, source)
}

func (m *Pomerium) goContainer(ctx context.Context, source *dagger.Directory) (*dagger.Container, error) {
	goVersion, err := toolVersion(ctx, m.Source, "golang")
	if err != nil {
		return nil, err
	}

	return dag.Container().
		From("golang:"+goVersion).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithEnvVariable("GOCACHE", "/root/.cache/go-build").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src"), nil
}

// toolVersion returns the version of tool listed in the source's .tool-versions.
func toolVersion(ctx context.Context, source *dagger.Directory, tool string) (string, error) {
	contents, err := source.File(".tool-versions").Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("reading .tool-versions: %w", err)
	}
	s := bufio.NewScanner(strings.NewReader(contents))
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) >= 2 && fields[0] == tool {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("%s not found in .tool-versions", tool)
}
