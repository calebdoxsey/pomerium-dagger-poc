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
	// The repository source: the workspace root, filtered by .gitignore.
	// +private
	Source *dagger.Directory

	// govulncheck version used by the govulncheck check.
	// +private
	GovulncheckVersion string
}

func New(
	// The current workspace, auto-populated by Dagger.
	ws *dagger.Workspace,
	// govulncheck version used by the govulncheck check.
	// +default="v1.5.0"
	govulncheckVersion string,
) *Pomerium {
	return &Pomerium{
		Source: ws.Directory("/", dagger.WorkspaceDirectoryOpts{
			Exclude:   []string{".git"},
			Gitignore: true,
		}),
		GovulncheckVersion: govulncheckVersion,
	}
}
