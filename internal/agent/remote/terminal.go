package remote

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/coder/websocket"

	"github.com/thomaspeterson/pc-inventory/internal/agent/transport"
	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

// ptySession ist die plattformabhängige PTY-Abstraktion. startPTY wird je Plattform
// (pty_unix.go / pty_windows.go / pty_other.go) bereitgestellt.
type ptySession interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Resize(cols, rows int) error
	Wait() int // blockiert bis Prozessende, liefert Exit-Code
	Close() error
}

// handleTerminal verbindet sich mit dem Server, startet eine PTY und pipet die
// Daten bidirektional: Binär-Frames = rohe Terminal-I/O, Text-Frames = Steuerung.
func handleTerminal(ctx context.Context, client *transport.Client, agentToken, session, shell, runas string, log *slog.Logger) {
	conn, err := client.DialTerminal(ctx, agentToken, session)
	if err != nil {
		log.Warn("terminal-ws fehlgeschlagen", "err", err)
		return
	}
	defer conn.CloseNow()

	pty, err := startPTY(shell, runas)
	if err != nil {
		_ = conn.Write(ctx, websocket.MessageBinary, []byte("Terminal konnte nicht gestartet werden: "+err.Error()+"\r\n"))
		writeExit(ctx, conn, -1)
		conn.Close(websocket.StatusInternalError, "pty start")
		return
	}
	defer pty.Close()

	ioCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Bricht die WS-Verbindung ab (Tab geschlossen) oder fällt eine Leserichtung
	// aus, killt das die Shell – pty.Wait() unten kehrt dann zurück.
	go func() {
		<-ioCtx.Done()
		_ = pty.Close()
	}()

	// PTY -> WS (Ausgabe als Binär-Frames).
	go func() {
		defer cancel()
		buf := make([]byte, 4096)
		for {
			n, rerr := pty.Read(buf)
			if n > 0 {
				if werr := conn.Write(ctx, websocket.MessageBinary, buf[:n]); werr != nil {
					return
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	// WS -> PTY (Binär = stdin, Text = Steuerung).
	go func() {
		defer cancel()
		for {
			typ, data, rerr := conn.Read(ctx)
			if rerr != nil {
				return
			}
			if typ == websocket.MessageBinary {
				_, _ = pty.Write(data)
				continue
			}
			var c shared.TermControl
			if json.Unmarshal(data, &c) == nil && c.Type == "resize" {
				_ = pty.Resize(c.Cols, c.Rows)
			}
		}
	}()

	// pty.Wait() liefert den maßgeblichen Exit-Code (kein Race mit den I/O-Goroutinen).
	code := pty.Wait()
	cancel()
	writeExit(context.Background(), conn, code)
	conn.Close(websocket.StatusNormalClosure, "exit")
	log.Info("terminal beendet", "session", session, "exit", code)
}

func writeExit(ctx context.Context, conn *websocket.Conn, code int) {
	msg, _ := json.Marshal(shared.TermControl{Type: "exit", Code: code})
	_ = conn.Write(ctx, websocket.MessageText, msg)
}
