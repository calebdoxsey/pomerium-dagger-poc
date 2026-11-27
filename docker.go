package main

import (
	"context"

	"dagger/pomerium/internal/dagger"
)

// DockerImage builds the repository's Dockerfile and returns the image.
//
// The engine's BuildKit runs the build, so its layers and the Dockerfile's
// cache mounts stay on the engine between runs.
//
// The source has no .git, so the binary's version and commit are empty.
func (m *Pomerium) DockerImage(
	ctx context.Context,
	// Platform to build, such as linux/arm64. Defaults to the engine's.
	// +optional
	platform dagger.Platform,
) (*dagger.Container, error) {
	if err := m.requireEngineVersion(ctx); err != nil {
		return nil, err
	}
	return m.Source.DockerBuild(dagger.DirectoryDockerBuildOpts{Platform: platform}), nil
}

// BuildDocker builds the repository's Dockerfile, without publishing the image.
//
// +check
func (m *Pomerium) BuildDocker(ctx context.Context) error {
	image, err := m.DockerImage(ctx, "")
	if err != nil {
		return err
	}
	_, err = image.Sync(ctx)
	return err
}
