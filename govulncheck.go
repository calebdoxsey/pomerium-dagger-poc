package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"dagger/pomerium/internal/dagger"
)

const vulnDBIndexURL = "https://vuln.go.dev/index/db.json"

// Govulncheck scans every module in the Go workspace for known vulnerabilities.
//
// The scan is cached until the source, the toolchain or the vulnerability
// database changes.
//
// +check
// +cache="never"
func (m *Pomerium) Govulncheck(ctx context.Context) error {
	ctr, err := m.GoBase(ctx)
	if err != nil {
		return err
	}
	_, err = ctr.
		WithExec([]string{"go", "install", "golang.org/x/vuln/cmd/govulncheck@" + govulncheckVersion}).
		// govulncheck downloads the vulnerability database at run time, which
		// Dagger can't see. Make the database version part of the cache key so
		// a cached result is never reused after new vulnerabilities are published.
		WithEnvVariable("GOVULNDB_MODIFIED", vulnDBModified(ctx)).
		// Make the Go GC work harder so the scan fits alongside other work on a
		// shared engine. On pomerium the peak drops from ~12GB to ~8.4GB for a few
		// seconds; lower limits thrash (6GiB: 7.9GB peak, 3x slower).
		WithEnvVariable("GOMEMLIMIT", "8GiB").
		With(func(c *dagger.Container) *dagger.Container {
			return withScript(c, "govulncheck.sh")
		}).
		Sync(ctx)
	return err
}

// vulnDBModified returns the last-modified time of the Go vulnerability
// database. If it can't be determined, it returns the current time so the
// scan runs uncached rather than reusing a possibly stale result.
func vulnDBModified(ctx context.Context) string {
	modified, err := fetchVulnDBModified(ctx)
	if err != nil {
		return "unknown-" + time.Now().UTC().Format(time.RFC3339Nano)
	}
	return modified
}

func fetchVulnDBModified(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vulnDBIndexURL, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status fetching %s: %s", vulnDBIndexURL, res.Status)
	}

	var index struct {
		Modified string `json:"modified"`
	}
	if err := json.NewDecoder(res.Body).Decode(&index); err != nil {
		return "", fmt.Errorf("decoding %s: %w", vulnDBIndexURL, err)
	}
	if index.Modified == "" {
		return "", fmt.Errorf("%s: missing modified time", vulnDBIndexURL)
	}
	return index.Modified, nil
}
