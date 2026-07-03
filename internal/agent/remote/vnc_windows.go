//go:build windows

package remote

import (
	"context"
	"fmt"

	"log/slog"
)

// startVNCServer wird für Windows in einem folgenden Schritt implementiert:
// UltraVNC (winvnc.exe) wird on-demand in der Konsolen-Session des angemeldeten
// Nutzers gestartet (CreateProcessAsUser, vgl. pty_windows_user.go), loopback +
// Einmalpasswort, optional QueryOnConnect für die Zustimmung.
func startVNCServer(_ context.Context, _ string, _ bool, _ *slog.Logger) (string, func(), error) {
	return "", nil, fmt.Errorf("fernsteuerung unter windows noch nicht implementiert")
}
