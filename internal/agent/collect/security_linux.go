//go:build linux

package collect

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// AVStatusJSON: klassische AV-Software ist unter Linux unüblich; wir melden das
// Vorhandensein von ClamAV, sonst „nicht zutreffend".
func AVStatusJSON(ctx context.Context) string {
	if have("clamscan") || have("clamdscan") {
		running := false
		if out := run(ctx, "sh", "-c", "pgrep -x clamd >/dev/null && echo yes"); strings.TrimSpace(out) == "yes" {
			running = true
		}
		return jsonOrError(AVStatus{Product: "ClamAV", Enabled: true, RealTime: running}, nil)
	}
	return jsonOrError(map[string]string{"info": "Kein klassischer Virenschutz unter Linux üblich"}, nil)
}

// BitLockerJSON: unter Linux nicht zutreffend.
func BitLockerJSON(ctx context.Context) string {
	return jsonOrError(map[string]any{"volumes": []BitLockerVolume{}, "info": "BitLocker nur unter Windows"}, nil)
}

// SmartJSON liest die Datenträgergesundheit via smartctl (falls installiert).
func SmartJSON(ctx context.Context) string {
	if !have("smartctl") {
		return jsonOrError(map[string]any{"disks": []SmartDisk{}, "info": "smartctl nicht installiert (Paket smartmontools)"}, nil)
	}
	scan := run(ctx, "smartctl", "--scan")
	var disks []SmartDisk
	for _, line := range nonEmptyLines(scan) {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		dev := f[0]
		out := run(ctx, "smartctl", "-H", "-i", dev)
		d := SmartDisk{Name: dev, Health: "Unbekannt"}
		for _, l := range nonEmptyLines(out) {
			low := strings.ToLower(l)
			if strings.Contains(low, "device model") || strings.Contains(low, "model number") {
				if i := strings.Index(l, ":"); i >= 0 {
					d.Model = strings.TrimSpace(l[i+1:])
				}
			}
			if strings.Contains(low, "smart overall-health") || strings.Contains(low, "smart health status") {
				if strings.Contains(low, "passed") || strings.Contains(low, "ok") {
					d.Health = "OK"
				} else if strings.Contains(low, "failed") {
					d.Health = "Fehler"
				}
				d.Detail = strings.TrimSpace(l)
			}
		}
		disks = append(disks, d)
	}
	return jsonOrError(map[string]any{"disks": disks}, nil)
}

// EventLogJSON liest die letzten journald-Einträge (optional nur ab Priorität).
func EventLogJSON(ctx context.Context, logName string, count int) string {
	if count <= 0 || count > 200 {
		count = 100
	}
	args := []string{"-n", itoa(count), "--no-pager", "-o", "json"}
	// logName als Priorität interpretieren: "errors" -> nur Warnungen/Fehler.
	if logName == "errors" {
		args = append(args, "-p", "warning")
	}
	cmd := exec.CommandContext(ctx, "journalctl", args...)
	out, err := cmd.Output()
	if err != nil {
		return jsonOrError(nil, err)
	}
	events := make([]EventEntry, 0, count)
	for _, line := range nonEmptyLines(string(out)) {
		var j struct {
			Ts       string `json:"__REALTIME_TIMESTAMP"`
			Msg      string `json:"MESSAGE"`
			Prio     string `json:"PRIORITY"`
			Unit     string `json:"_SYSTEMD_UNIT"`
			Comm     string `json:"_COMM"`
		}
		if json.Unmarshal([]byte(line), &j) != nil {
			continue
		}
		src := j.Unit
		if src == "" {
			src = j.Comm
		}
		msg := j.Msg
		if len(msg) > 300 {
			msg = msg[:300]
		}
		events = append(events, EventEntry{Time: journaldTime(j.Ts), Level: prioLabel(j.Prio), Source: src, Message: msg})
	}
	b, _ := json.Marshal(map[string]any{"events": events})
	return string(b)
}

// journaldTime wandelt __REALTIME_TIMESTAMP (Mikrosekunden) in ISO-Zeit.
func journaldTime(us string) string {
	n, err := strconv.ParseInt(us, 10, 64)
	if err != nil || n == 0 {
		return ""
	}
	return time.Unix(0, n*1000).UTC().Format(time.RFC3339)
}

func prioLabel(p string) string {
	switch p {
	case "0", "1", "2", "3":
		return "Error"
	case "4":
		return "Warning"
	case "5", "6":
		return "Information"
	default:
		return "Debug"
	}
}
