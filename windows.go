package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// windowsArchitectures are the Windows architectures BuildWindows builds for.
var windowsArchitectures = []string{"amd64", "arm64"}

// BuildWindows cross-compiles the repository for Windows (amd64 and arm64),
// except Unix-only server packages and anything importing them.
//
// +check
func (m *Pomerium) BuildWindows(ctx context.Context) error {
	// The Unix-only packages excluded from the build are the only ones that
	// embed the envoy binaries, so this skips Prepare.
	ctr, err := m.goContainer(ctx, m.Source)
	if err != nil {
		return err
	}
	eg, ctx := errgroup.WithContext(ctx)
	for _, arch := range windowsArchitectures {
		eg.Go(func() error {
			_, err := withScript(ctr.
				WithEnvVariable("GOOS", "windows").
				WithEnvVariable("GOARCH", arch),
				"build-windows.sh").
				Sync(ctx)
			if err != nil {
				return fmt.Errorf("windows/%s: %w", arch, err)
			}
			return nil
		})
	}
	return eg.Wait()
}
