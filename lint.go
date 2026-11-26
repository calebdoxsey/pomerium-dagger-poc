package main

import (
	"context"
)

// Lint runs golangci-lint on the repository.
//
// The golangci-lint version is golangciLintVersion.
//
// +check
func (m *Pomerium) Lint(ctx context.Context) error {
	ctr, err := m.GoBase(ctx)
	if err != nil {
		return err
	}
	// Use the official release binary, as golangci-lint recommends, but run it
	// in the Go container so it uses the repository's Go version. (The image's
	// own Go toolchain can be older than the one go.mod requires.)
	golangciLint := dag.Container().
		From("golangci/golangci-lint:" + golangciLintVersion).
		File("/usr/bin/golangci-lint")
	_, err = ctr.
		WithFile("/usr/local/bin/golangci-lint", golangciLint).
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint")).
		WithEnvVariable("GOLANGCI_LINT_CACHE", "/root/.cache/golangci-lint").
		WithExec([]string{"golangci-lint", "run", "--timeout=10m"}).
		Sync(ctx)
	return err
}
