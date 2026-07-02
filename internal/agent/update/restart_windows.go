//go:build windows

package update

import (
	"os/exec"
	"syscall"
)

// serviceName muss zum Dienstnamen in cmd/agent passen.
const serviceName = "pc-inventory-agent"

// restart startet einen losgelösten Helfer, der nach kurzer Wartezeit den Dienst
// neu startet – dadurch lädt der Service Control Manager das neue Binary.
// (Ein laufendes .exe lässt sich unter Windows nicht im selben Prozess re-exec'en;
// der exe-Pfad wird daher nicht benötigt.)
func restart(_ string) error {
	const (
		detachedProcess     = 0x00000008
		createNewProcessGrp = 0x00000200
	)
	cmd := exec.Command("cmd", "/c",
		"timeout /t 2 /nobreak >nul & sc stop "+serviceName+" & sc start "+serviceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: detachedProcess | createNewProcessGrp}
	return cmd.Start()
}
