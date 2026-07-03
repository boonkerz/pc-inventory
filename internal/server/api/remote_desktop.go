package api

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"

	"github.com/thomaspeterson/pc-inventory/internal/server/model"
	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

// Web-Fernsteuerung (Remote Desktop) über denselben On-demand-Tunnel wie das
// Terminal: der Browser rendert mit noVNC, am Endpunkt läuft on-demand ein nativer
// VNC-Server (127.0.0.1), unser Server reicht nur die RFB-Bytes durch.
//
//  1. Der Browser fordert per POST /remote/start eine Sitzung an. Der Server
//     erzeugt Session-Token + Einmalpasswort, weckt den Agent ("open_vnc") und
//     gibt {session, password} zurück.
//  2. Der Agent startet den VNC-Server mit dem Passwort und meldet sich mit einer
//     WebSocket für die Session (handleAgentTerminal – wiederverwendet).
//  3. Der Browser öffnet die noVNC-WS /remote/ws?session=…; der Server paart
//     Browser- und Agent-WS und relayed die Frames (relay – wiederverwendet).

const remoteSessionTTL = 60 * time.Second // so lange wartet eine Sitzung auf die Browser-WS

// vncPassword erzeugt ein 8-Zeichen-Einmalpasswort (RFB-Auth ist auf 8 Zeichen
// begrenzt) aus einem verwechslungsarmen Alphabet.
func vncPassword() string {
	const alpha = "abcdefghijkmnpqrstuvwxyz23456789"
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alpha[int(b[i])%len(alpha)]
	}
	return string(b)
}

type remoteStartResponse struct {
	Session  string `json:"session"`
	Password string `json:"password"`
}

// handleRemoteStart erzeugt eine Remote-Desktop-Sitzung und weckt den Agent.
func (s *Server) handleRemoteStart(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	sessID := newSessionID()
	pass := vncPassword()

	sess := &termSession{
		deviceID:  deviceID,
		agentConn: make(chan *websocket.Conn, 1),
		done:      make(chan struct{}),
	}
	s.term.addSession(sessID, sess)

	// Sitzung aufräumen, falls die Browser-WS nie kommt (kein Leak).
	go func() {
		select {
		case <-sess.done:
		case <-time.After(remoteSessionTTL):
			s.term.takeSession(sessID)
		}
	}()

	consent := s.resolveRemoteConsent(r.Context(), deviceID)
	s.term.requestWake(deviceID, shared.WaitResponse{
		Type: "open_vnc", Session: sessID, Password: pass, Consent: consent,
	})

	var uname string
	if u := userFrom(r.Context()); u != nil {
		uname = u.Username
	}
	_ = s.store.InsertAudit(r.Context(), model.AuditEntry{
		TS: time.Now().UTC(), Username: uname, Action: "Fernsteuerung gestartet",
		Method: r.Method, Path: "/devices/" + deviceID, Status: http.StatusOK,
	})
	s.writeJSON(w, http.StatusOK, remoteStartResponse{Session: sessID, Password: pass})
}

// resolveRemoteConsent ermittelt, ob der Nutzer am Gerät die Fernsteuerung
// bestätigen muss. Phase 1: standardmäßig unbeaufsichtigt (false); die
// zielgruppenabhängige Auflösung (device→site→client) folgt in Phase 3.
func (s *Server) resolveRemoteConsent(ctx context.Context, deviceID string) bool {
	return false
}

// handleDeviceVNC nimmt die noVNC-WS des Browsers entgegen und relayed sie an die
// (bereits per /remote/start angeforderte) Agent-Session.
func (s *Server) handleDeviceVNC(w http.ResponseWriter, r *http.Request) {
	sessID := r.URL.Query().Get("session")
	s.term.mu.Lock()
	sess := s.term.session[sessID]
	s.term.mu.Unlock()
	if sess == nil || sess.deviceID != chi.URLParam(r, "id") {
		s.writeErr(w, http.StatusNotFound, "unbekannte session")
		return
	}
	defer s.term.takeSession(sessID)

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		return
	}
	c.SetReadLimit(4 << 20) // RFB-Frames können größer sein als Terminal-I/O
	defer c.CloseNow()

	ctx := r.Context()
	var agent *websocket.Conn
	select {
	case agent = <-sess.agentConn:
	case <-time.After(agentConnectWindow):
		c.Close(websocket.StatusTryAgainLater, "agent nicht erreichbar")
		return
	case <-ctx.Done():
		return
	}

	relay(ctx, c, agent)
	close(sess.done)
}
