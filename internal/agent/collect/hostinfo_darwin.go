//go:build darwin

package collect

import (
	"context"
	"os/exec"
	"strings"
)

// hardwareIdentity liest die Hardware-Identität aus ioreg.
func hardwareIdentity(ctx context.Context) (vendor, model, serial string) {
	vendor = "Apple"
	out, err := exec.CommandContext(ctx, "ioreg", "-d2", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return vendor, "", ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.Contains(line, "\"model\""):
			model = ioregValue(line)
		case strings.Contains(line, "IOPlatformSerialNumber"):
			serial = ioregValue(line)
		}
	}
	return vendor, model, serial
}

func ioregValue(line string) string {
	if i := strings.LastIndex(line, "= "); i >= 0 {
		return strings.Trim(strings.TrimSpace(line[i+2:]), "\"<>")
	}
	return ""
}
