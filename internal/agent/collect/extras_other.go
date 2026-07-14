//go:build !linux && !windows && !darwin

package collect

import (
	"context"

	"github.com/boonkerz/roster/internal/shared"
)

func osExtras(_ context.Context) ([]shared.SoftwarePackage, []shared.Printer, []string) {
	return nil, nil, nil
}
