package collect

import (
	"encoding/json"
	"sort"
)

// ServiceInfo beschreibt einen Systemdienst (Windows-Dienst / systemd-Unit).
type ServiceInfo struct {
	Name      string `json:"name"`
	Display   string `json:"display,omitempty"`
	Running   bool   `json:"running"`
	StartType string `json:"start_type,omitempty"`
}

// servicesJSON serialisiert die Dienstliste (oder einen Fehler) nach JSON.
func servicesJSON(list []ServiceInfo, err error) string {
	if err != nil {
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(b)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	b, _ := json.Marshal(map[string]any{"services": list})
	return string(b)
}
