//go:build !linux && !windows && !darwin

package collect

import (
	"context"

	"github.com/boonkerz/roster/internal/shared"
)

func hwExtras(_ context.Context) ([]string, []shared.PhysicalDisk) {
	return nil, nil
}
