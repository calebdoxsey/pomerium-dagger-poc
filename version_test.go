package main

import (
	"os"
	"regexp"
	"testing"
)

// The GitHub Action installs the Dagger CLI; it must install the version the
// module requires.
func TestActionVersionMatchesDaggerVersion(t *testing.T) {
	const path = "actions/run/action.yml"
	action, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile(`(?m)^ +version: (\S+)$`).FindAllSubmatch(action, -1)
	if len(matches) != 1 {
		t.Fatalf("%s: expected one dagger version, found %d", path, len(matches))
	}
	if got := string(matches[0][1]); got != daggerVersion {
		t.Errorf("%s installs dagger %s, but daggerVersion is %s", path, got, daggerVersion)
	}
}
