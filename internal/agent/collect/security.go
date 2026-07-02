package collect

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func trimSpace(s string) string { return strings.TrimSpace(s) }

var errUnparsable = errors.New("Ausgabe nicht lesbar")

func itoa(n int) string { return strconv.Itoa(n) }

// EventEntry ist ein Log-Ereignis (Windows-Eventlog oder journald).
type EventEntry struct {
	Time    string `json:"time"`
	ID      int    `json:"id,omitempty"`
	Level   string `json:"level"`
	Source  string `json:"source,omitempty"`
	Message string `json:"message"`
}

// wrapEvents normalisiert die Windows-Eventlog-JSON (t,id,lvl,src,msg) in EventEntry.
func wrapEvents(raw []byte) string {
	type row struct {
		T   string `json:"t"`
		ID  int    `json:"id"`
		Lvl string `json:"lvl"`
		Src string `json:"src"`
		Msg string `json:"msg"`
	}
	var rows []row
	trimmed := trimSpace(string(raw))
	if len(trimmed) > 0 && trimmed[0] == '[' {
		_ = json.Unmarshal(raw, &rows)
	} else if trimmed != "" {
		var one row
		if json.Unmarshal(raw, &one) == nil {
			rows = []row{one}
		}
	}
	events := make([]EventEntry, 0, len(rows))
	for _, r := range rows {
		events = append(events, EventEntry{Time: r.T, ID: r.ID, Level: r.Lvl, Source: r.Src, Message: r.Msg})
	}
	b, _ := json.Marshal(map[string]any{"events": events})
	return string(b)
}

// AVStatus beschreibt den Virenschutz-Status.
type AVStatus struct {
	Product      string `json:"product"`
	Enabled      bool   `json:"enabled"`
	RealTime     bool   `json:"realtime"`
	SignatureAge int    `json:"signature_age_days"`
	Version      string `json:"version,omitempty"`
}

// BitLockerVolume beschreibt den Verschlüsselungsstatus eines Volumes.
type BitLockerVolume struct {
	MountPoint   string `json:"mount_point"`
	Protection   string `json:"protection"` // On | Off | Unknown
	Percent      int    `json:"percent"`
	RecoveryKey  string `json:"recovery_key,omitempty"`
	RecoveryID   string `json:"recovery_id,omitempty"`
}

// SmartDisk beschreibt die Gesundheit eines physischen Datenträgers.
type SmartDisk struct {
	Name   string `json:"name"`
	Model  string `json:"model,omitempty"`
	Health string `json:"health"` // OK | Warnung | Fehler | Unbekannt
	Detail string `json:"detail,omitempty"`
}

func jsonOrError(v any, err error) string {
	if err != nil {
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(b)
	}
	b, _ := json.Marshal(v)
	return string(b)
}
