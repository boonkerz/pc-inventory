//go:build !linux && !windows

package remote

import (
	"context"
	"fmt"

	"log/slog"
)

// startVNCServer ist auf dieser Plattform (noch) nicht unterstützt (z.B. macOS →
// eingebautes Screen Sharing, Phase 4).
func startVNCServer(_ context.Context, _ string, _ bool, _ *slog.Logger) (string, func(), error) {
	return "", nil, fmt.Errorf("fernsteuerung auf dieser plattform nicht unterstützt")
}
