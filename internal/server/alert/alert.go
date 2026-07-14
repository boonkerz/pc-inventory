// Package alert versendet Benachrichtigungen über modulare Kanäle (Provider).
package alert

import (
	"context"
	"log/slog"
	"time"

	"github.com/boonkerz/roster/internal/server/model"
)

// Dispatch versendet eine Benachrichtigung asynchron über einen Kanal. Der Versand
// ist vom Request-Lebenszyklus entkoppelt (eigener Kontext). Geltungsbereich und
// Schweregrad-Filterung passieren beim Aufrufer (siehe alertFailing).
func Dispatch(log *slog.Logger, ch model.AlertChannel, n Notification) {
	p := registry[ch.Type]
	if p == nil {
		log.Warn("unbekannter alert-kanal-typ", "type", ch.Type)
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := p.Send(ctx, ch.Config, n); err != nil {
			log.Warn("alert-kanal fehlgeschlagen", "type", ch.Type, "name", ch.Name, "err", err)
		}
	}()
}
