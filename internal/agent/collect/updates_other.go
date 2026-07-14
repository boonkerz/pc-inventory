//go:build !linux && !windows && !darwin

package collect

import (
	"context"

	"github.com/boonkerz/roster/internal/shared"
)

func osUpdateList(_ context.Context) ([]shared.UpdateItem, bool) { return nil, false }
