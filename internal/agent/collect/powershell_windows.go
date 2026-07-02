//go:build windows

package collect

import (
	"bytes"
	"context"
	"os/exec"
)

// psOutput führt ein PowerShell-Skript aus und liefert dessen stdout als UTF-8.
// Windows-PowerShell gibt sonst in der OEM-Codepage aus (kaputte Umlaute); der Prefix
// erzwingt UTF-8 ohne BOM, ein evtl. dennoch vorhandener BOM wird entfernt.
func psOutput(ctx context.Context, script string) ([]byte, error) {
	const prefix = "[Console]::OutputEncoding=(New-Object System.Text.UTF8Encoding $false); "
	out, err := exec.CommandContext(ctx, "powershell",
		"-NoProfile", "-NonInteractive", "-Command", prefix+script).Output()
	out = bytes.TrimPrefix(out, []byte{0xEF, 0xBB, 0xBF})
	return out, err
}
