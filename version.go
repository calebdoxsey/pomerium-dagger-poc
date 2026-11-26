package main

import (
	"context"
	"fmt"
	"strings"
)

// requireEngineVersion fails unless the Dagger engine running this module is
// the version pinned by the "dagger" entry in the repository's .tool-versions.
//
// Dagger accepts mismatched CLI and engine versions, so without this a stale
// CLI (which starts an engine of its own version) or an out-of-date shared CI
// engine would run silently.
func (m *Pomerium) requireEngineVersion(ctx context.Context) error {
	want, err := toolVersion(ctx, m.Source, "dagger")
	if err != nil {
		return err
	}
	got, err := dag.Version(ctx)
	if err != nil {
		return fmt.Errorf("getting dagger engine version: %w", err)
	}
	if normalizeVersion(got) != normalizeVersion(want) {
		return fmt.Errorf("dagger engine is %s, but .tool-versions requires %s: "+
			"install the dagger CLI from .tool-versions, or point DAGGER_ENGINE at an engine of that version",
			got, want)
	}
	return nil
}

// normalizeVersion strips the leading "v" and any build metadata, so
// "v1.0.0-beta.30+0b5f50da" and "1.0.0-beta.30" compare equal.
func normalizeVersion(v string) string {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	v, _, _ = strings.Cut(v, "+")
	return v
}
