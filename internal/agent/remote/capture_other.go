//go:build !windows

package remote

import "log/slog"

// newScreenSource: echte Bildschirmaufnahme unter Linux/macOS folgt; vorerst ein
// Testbild (Übertragung/Tunnel sind damit verifizierbar).
func newScreenSource(_ *slog.Logger) (screenSource, error) {
	return newSyntheticSource(), nil
}
