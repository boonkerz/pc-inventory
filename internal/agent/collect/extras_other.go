//go:build !linux && !windows && !darwin

package collect

import (
	"context"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

func osExtras(_ context.Context) ([]shared.SoftwarePackage, []shared.Printer, []string) {
	return nil, nil, nil
}
