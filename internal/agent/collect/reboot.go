package collect

import (
	"context"
	"os/exec"
	"runtime"
)

// Reboot startet das System neu (mit den Rechten des Agent-Dienstes).
func Reboot(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "shutdown", "/r", "/t", "0").Run()
	}
	return exec.CommandContext(ctx, "shutdown", "-r", "now").Run()
}
