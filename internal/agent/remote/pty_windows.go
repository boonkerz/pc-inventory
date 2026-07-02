//go:build windows

package remote

import (
	"context"
	"fmt"
	"sync"

	"github.com/UserExistsError/conpty"
)

// winPTY führt cmd.exe bzw. powershell.exe unter einem ConPTY (Pseudo-Konsole) aus.
type winPTY struct {
	cpty      *conpty.ConPty
	closeOnce sync.Once
}

// startPTY startet die gewünschte Shell. runas=="user" führt sie im Kontext des
// aktuell an der Konsole angemeldeten Benutzers aus (CreateProcessAsUser mit dem
// Token der aktiven Sitzung – siehe pty_windows_user.go; setzt voraus, dass der
// Agent als SYSTEM-Dienst läuft). Sonst läuft die Shell im Dienst-Kontext (SYSTEM).
func startPTY(shell, runas string) (ptySession, error) {
	if runas == "user" {
		return startUserPTY(shell)
	}
	if !conpty.IsConPtyAvailable() {
		return nil, fmt.Errorf("ConPTY auf diesem Windows nicht verfügbar")
	}
	cpty, err := conpty.Start(windowsShellCmd(shell), conpty.ConPtyDimensions(120, 30))
	if err != nil {
		return nil, err
	}
	return &winPTY{cpty: cpty}, nil
}

// windowsShellCmd liefert die zu startende Kommandozeile für die gewählte Shell.
func windowsShellCmd(shell string) string {
	if shell == "powershell" {
		return "powershell.exe -NoLogo"
	}
	return "cmd.exe"
}

func (p *winPTY) Read(b []byte) (int, error)  { return p.cpty.Read(b) }
func (p *winPTY) Write(b []byte) (int, error) { return p.cpty.Write(b) }

func (p *winPTY) Resize(cols, rows int) error { return p.cpty.Resize(cols, rows) }

func (p *winPTY) Wait() int {
	code, err := p.cpty.Wait(context.Background())
	if err != nil {
		return -1
	}
	return int(code)
}

// Close ist idempotent: Wait() und der Teardown-Pfad können es nebenläufig aufrufen.
func (p *winPTY) Close() error {
	p.closeOnce.Do(func() { _ = p.cpty.Close() })
	return nil
}
