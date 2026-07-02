//go:build linux

package collect

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ListServices listet systemd-Dienst-Units (Name, Beschreibung, laufend).
func ListServices(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "systemctl",
		"list-units", "--type=service", "--all", "--no-legend", "--plain", "--no-pager").Output()
	if err != nil {
		return servicesJSON(nil, err)
	}
	var list []ServiceInfo
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if len(f) < 4 {
			continue
		}
		// Spalten: UNIT LOAD ACTIVE SUB DESCRIPTION…
		display := ""
		if len(f) > 4 {
			display = strings.Join(f[4:], " ")
		}
		list = append(list, ServiceInfo{Name: f[0], Display: display, Running: f[3] == "running"})
	}
	return servicesJSON(list, nil)
}

// ControlService startet/stoppt/startet einen Dienst neu (systemctl).
func ControlService(ctx context.Context, name, action string) (int, string) {
	switch action {
	case "start", "stop", "restart":
	default:
		return 1, "ungültige Aktion"
	}
	if strings.ContainsAny(name, " ;&|$`\n") {
		return 1, "ungültiger Dienstname"
	}
	out, err := exec.CommandContext(ctx, "systemctl", action, name).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return 1, msg
	}
	return 0, fmt.Sprintf("%s: %s ausgeführt", name, action)
}
