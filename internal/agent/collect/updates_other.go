//go:build !linux && !windows && !darwin

package collect

import (
	"context"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

func osUpdateList(_ context.Context) ([]shared.UpdateItem, bool) { return nil, false }
