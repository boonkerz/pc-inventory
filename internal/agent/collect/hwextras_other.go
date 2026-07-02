//go:build !linux && !windows && !darwin

package collect

import (
	"context"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

func hwExtras(_ context.Context) ([]string, []shared.PhysicalDisk) {
	return nil, nil
}
