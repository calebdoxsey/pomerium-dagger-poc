package main

import (
	"context"
)

// BuildDocker builds the repository's Dockerfile, without publishing the image.
//
// The engine's BuildKit runs the build, so its layers and the Dockerfile's
// cache mounts stay on the engine between runs.
//
// +check
func (m *Pomerium) BuildDocker(ctx context.Context) error {
	if err := m.requireEngineVersion(ctx); err != nil {
		return err
	}
	// The source has no .git, so the binary's version and commit are empty;
	// that's fine for an image that's thrown away.
	_, err := m.Source.DockerBuild().Sync(ctx)
	return err
}
