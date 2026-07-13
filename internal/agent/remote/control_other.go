//go:build !windows

package remote

import "log/slog"

// Steuerkanal-Aktionen (Strg+Alt+Entf, Meldung) sind nur unter Windows umgesetzt.

func agentSendSAS(log *slog.Logger)                  {}
func agentShowMessage(log *slog.Logger, text string) {}
