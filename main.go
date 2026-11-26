// Shared CI/build functions for Pomerium repositories.
//
// Each Pomerium repository installs this module into its Dagger workspace
// (dagger.toml) and runs its checks with `dagger check`, both locally and in CI.
package main

import (
	"dagger/pomerium/internal/dagger"
)

// Pomerium is the root object of the module.
type Pomerium struct {
	// The repository source: the workspace root, minus sourceExclude.
	// +private
	Source *dagger.Directory
}

// sourceExclude lists large local-only paths to leave out of the source.
//
// .gitignore isn't applied: Dagger filters by the patterns alone, which also
// drops tracked files that match them (in pomerium, .golangci.yml and dozens of
// test fixtures). A CI checkout only has tracked files, so there this is the
// exact checkout; locally, untracked files are included too.
var sourceExclude = []string{
	".git",
	"**/node_modules",
	"bin",
	// downloaded by Prepare
	"pkg/envoy/files/envoy-*",
}

func New(
	// The current workspace, auto-populated by Dagger.
	ws *dagger.Workspace,
) *Pomerium {
	return &Pomerium{
		Source: ws.Directory("/", dagger.WorkspaceDirectoryOpts{Exclude: sourceExclude}),
	}
}
