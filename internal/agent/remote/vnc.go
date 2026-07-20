package remote

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/coder/websocket"

	"github.com/boonkerz/roster/internal/agent/transport"
)

// resolutionSetter wird von Bildschirmquellen implementiert, die die physische
// Auflösung des Geräts anpassen können (adaptive Auflösung). Die Umstellung MUSS in
// der Nutzer-Session laufen (Session-0-Isolation), daher als Quellen-Fähigkeit über
// den Aufnahme-Helfer geroutet (Windows). w<=0||h<=0 = ursprüngliche wiederherstellen.
type resolutionSetter interface {
	SetResolution(w, h int)
}

// handleVNC bedient eine Fernsteuerungs-Sitzung: der Agent ist selbst der VNC-Server.
// Er verbindet sich per WebSocket mit dem Server und fährt darauf den eingebauten
// RFB-Server (rfb.go), der den Bildschirm aufnimmt und Eingaben umsetzt. Keine
// Fremdsoftware.
func handleVNC(ctx context.Context, client *transport.Client, agentToken, session, _ string, consent bool, monitor int, log *slog.Logger) {
	conn, err := client.DialTerminal(ctx, agentToken, session)
	if err != nil {
		log.Warn("vnc-websocket fehlgeschlagen", "err", err)
		return
	}
	defer conn.CloseNow()

	// Zielgruppenabhängige Zustimmung: der angemeldete Nutzer muss bestätigen.
	if consent && !confirmRemote(log) {
		log.Info("fernsteuerung am gerät abgelehnt/keine antwort")
		conn.Close(websocket.StatusPolicyViolation, "am gerät abgelehnt")
		return
	}

	src, err := newResilientSource(log, monitor)
	if err != nil {
		log.Warn("bildschirmaufnahme nicht verfügbar", "err", err)
		conn.Close(websocket.StatusInternalError, "keine aufnahme")
		return
	}
	defer src.Close()

	// Adaptive Auflösung: beim Sitzungsende die native Auflösung wiederherstellen
	// (läuft VOR src.Close, solange der Helfer noch lebt). resilientSource leitet an
	// die innere Quelle weiter bzw. macht No-op, wenn sie es nicht kann.
	defer src.SetResolution(0, 0)

	w, h := src.Bounds()
	log.Info("fernsteuerung: rfb-server startet", "size", fmt.Sprintf("%dx%d", w, h))
	nc := websocket.NetConn(ctx, conn, websocket.MessageBinary)
	if err := rfbServe(ctx, nc, src, log); err != nil && ctx.Err() == nil {
		log.Debug("rfb-server beendet", "err", err)
	}
	conn.Close(websocket.StatusNormalClosure, "ende")
	log.Info("fernsteuerung beendet", "session", session)
}
