//go:build !windows

package remote

import "log/slog"

// confirmRemote: auf Nicht-Windows-Plattformen (derzeit view-only) keine Nachfrage.
func confirmRemote(_ *slog.Logger) bool { return true }
