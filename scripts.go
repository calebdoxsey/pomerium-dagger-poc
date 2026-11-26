package main

import (
	"embed"

	"dagger/pomerium/internal/dagger"
)

// scripts holds the shell scripts that checks run in containers. They are
// separate files so they can be shellchecked.
//
//go:embed scripts/*.sh
var scripts embed.FS

// withScript runs scripts/<name> with bash in ctr, passing args to it.
func withScript(ctr *dagger.Container, name string, args ...string) *dagger.Container {
	contents, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		// The scripts are embedded, so a missing one is a bug in this module.
		panic(err)
	}
	path := "/usr/local/share/pomerium-dagger/" + name
	return ctr.
		WithNewFile(path, string(contents)).
		WithExec(append([]string{"bash", path}, args...))
}
