package remote

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/coder/websocket"

	"github.com/boonkerz/roster/internal/agent/transport"
)

// resController ändert die physische Bildschirmauflösung des Geräts für die Dauer
// einer Fernsteuerungs-Sitzung (adaptive Auflösung: der Viewer bittet um eine zum
// Fenster passende Größe) und stellt beim Sitzungsende die ursprüngliche wieder her.
// Nur Windows setzt das echt um; sonst No-op (siehe resolution_*.go).
type resController interface {
	Set(w, h int) // gewünschte Größe; w<=0||h<=0 = nativ wiederherstellen
	Restore()     // ursprüngliche Auflösung zurücksetzen (Sitzungsende)
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

	// Adaptive Auflösung: der Viewer kann eine zum Fenster passende Bildschirmgröße
	// anfordern; beim Sitzungsende wird die ursprüngliche wiederhergestellt.
	res := newResController(log)
	defer res.Restore()

	w, h := src.Bounds()
	log.Info("fernsteuerung: rfb-server startet", "size", fmt.Sprintf("%dx%d", w, h))
	nc := websocket.NetConn(ctx, conn, websocket.MessageBinary)
	if err := rfbServe(ctx, nc, src, res, log); err != nil && ctx.Err() == nil {
		log.Debug("rfb-server beendet", "err", err)
	}
	conn.Close(websocket.StatusNormalClosure, "ende")
	log.Info("fernsteuerung beendet", "session", session)
}
