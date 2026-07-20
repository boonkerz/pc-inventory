//go:build !windows

package remote

import "log/slog"

// Auf Nicht-Windows-Plattformen ist die Bildschirmaufnahme derzeit synthetisch bzw.
// eine physische Auflösungsänderung nicht vorgesehen – adaptive Auflösung ist ein
// No-op. (Echte Linux-Aufnahme/Wayland folgt separat.)

type nopResController struct{}

func newResController(_ *slog.Logger) resController { return nopResController{} }

func (nopResController) Set(_, _ int) {}
func (nopResController) Restore()     {}
