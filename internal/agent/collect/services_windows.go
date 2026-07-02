//go:build windows

package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ListServices listet Windows-Dienste via PowerShell (Get-Service).
func ListServices(ctx context.Context) string {
	out, err := psOutput(ctx, `Get-Service | Select-Object Name,DisplayName,@{N='Status';E={"$($_.Status)"}},@{N='StartType';E={"$($_.StartType)"}} | ConvertTo-Json -Compress`)
	if err != nil {
		return servicesJSON(nil, err)
	}
	type row struct {
		Name        string `json:"Name"`
		DisplayName string `json:"DisplayName"`
		Status      string `json:"Status"`
		StartType   string `json:"StartType"`
	}
	trimmed := strings.TrimSpace(string(out))
	var rows []row
	if strings.HasPrefix(trimmed, "[") {
		_ = json.Unmarshal(out, &rows)
	} else if trimmed != "" {
		var single row
		if json.Unmarshal(out, &single) == nil {
			rows = []row{single}
		}
	}
	list := make([]ServiceInfo, 0, len(rows))
	for _, r := range rows {
		list = append(list, ServiceInfo{
			Name: r.Name, Display: r.DisplayName,
			Running: strings.EqualFold(r.Status, "Running"), StartType: r.StartType,
		})
	}
	return servicesJSON(list, nil)
}

// ControlService startet/stoppt/startet einen Windows-Dienst neu.
func ControlService(ctx context.Context, name, action string) (int, string) {
	var cmd string
	switch action {
	case "start":
		cmd = "Start-Service"
	case "stop":
		cmd = "Stop-Service"
	case "restart":
		cmd = "Restart-Service"
	default:
		return 1, "ungültige Aktion"
	}
	if strings.ContainsAny(name, "'\";`\r\n") {
		return 1, "ungültiger Dienstname"
	}
	out, err := psOutput(ctx, fmt.Sprintf("%s -Name '%s' -ErrorAction Stop", cmd, name))
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return 1, msg
	}
	return 0, fmt.Sprintf("%s: %s ausgeführt", name, action)
}
