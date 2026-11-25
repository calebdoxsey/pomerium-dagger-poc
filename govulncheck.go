package main

import (
	"context"
)

// Govulncheck scans every module in the Go workspace for known vulnerabilities.
//
// +check
func (m *Pomerium) Govulncheck(ctx context.Context) error {
	ctr, err := m.GoBase(ctx)
	if err != nil {
		return err
	}
	_, err = ctr.
		WithExec([]string{"go", "install", "golang.org/x/vuln/cmd/govulncheck@" + m.GovulncheckVersion}).
		WithExec([]string{"sh", "-euc", `
for dir in $(go list -m -f '{{.Dir}}'); do
	echo "==> govulncheck $dir"
	(cd "$dir" && govulncheck ./...)
done
`}).
		Sync(ctx)
	return err
}
