package collect

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DUEntry ist ein direkter Eintrag eines Verzeichnisses mit (rekursiver) Größe.
type DUEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Dir      bool   `json:"dir"`
	Items    int64  `json:"items,omitempty"`    // Dateien darunter (nur Ordner)
	Counting bool   `json:"counting,omitempty"` // wird gerade noch durchgezählt
}

// DUResult ist das Ergebnis eines Verzeichnis-Scans (eine Ebene).
type DUResult struct {
	Path    string    `json:"path"`
	Parent  string    `json:"parent"`
	Entries []DUEntry `json:"entries"`
	Error   string    `json:"error,omitempty"`
}

const duMaxEntries = 1000 // Sicherheitslimit gegen riesige Verzeichnisse

// DirUsage zählt die direkten Einträge von path durch und liefert das fertige
// Ergebnis als JSON (ohne Zwischenstände – für Tests/Einmalnutzung).
func DirUsage(ctx context.Context, path string) string {
	var final DUResult
	DirUsageStream(ctx, path, func(r DUResult) { final = r })
	out, _ := json.Marshal(final)
	return string(out)
}

// DirUsageStream zählt die direkten Einträge von path durch und ruft emit mit
// fortlaufenden Zwischenständen auf: zuerst sofort mit allen Namen (Ordner als
// „counting"), dann ~alle 700 ms mit hochgezählten Größen, am Ende sortiert. So
// sieht der Aufrufer die Unterordner gleich und die Zahlen live wachsen. Bricht
// bei ctx-Abbruch ab; nicht lesbare Pfade werden übersprungen.
func DirUsageStream(ctx context.Context, path string, emit func(DUResult)) {
	res := DUResult{Path: path, Parent: filepath.Dir(path)}
	if res.Parent == path {
		res.Parent = "" // Wurzel hat kein „darüber"
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		res.Error = err.Error()
		emit(res)
		return
	}
	// 1) Alle Namen sofort aufnehmen: Dateien mit Größe, Ordner als „counting".
	var dirIdx []int
	for _, e := range entries {
		full := filepath.Join(path, e.Name())
		if e.IsDir() {
			res.Entries = append(res.Entries, DUEntry{Name: e.Name(), Path: full, Dir: true, Counting: true})
			dirIdx = append(dirIdx, len(res.Entries)-1)
		} else {
			var sz int64
			if info, err := e.Info(); err == nil {
				sz = info.Size()
			}
			res.Entries = append(res.Entries, DUEntry{Name: e.Name(), Path: full, Size: sz})
		}
	}
	emitCopy := func() {
		cp := res
		cp.Entries = append([]DUEntry(nil), res.Entries...)
		emit(cp)
	}
	last := time.Now()
	throttle := func() {
		if time.Since(last) >= 700*time.Millisecond {
			last = time.Now()
			emitCopy()
		}
	}
	emitCopy() // sofort: Unterordner sichtbar

	// 2) Ordner nacheinander durchzählen, dabei laufend Zwischenstände melden.
	for _, idx := range dirIdx {
		if ctx.Err() != nil {
			res.Error = "Scan abgebrochen"
			break
		}
		var total, count int64
		_ = filepath.WalkDir(res.Entries[idx].Path, func(_ string, d os.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil || d.IsDir() {
				return nil
			}
			if info, e := d.Info(); e == nil && info.Mode().IsRegular() {
				total += info.Size()
				count++
				if count%2000 == 0 {
					res.Entries[idx].Size = total
					res.Entries[idx].Items = count
					throttle()
				}
			}
			return nil
		})
		res.Entries[idx].Size = total
		res.Entries[idx].Items = count
		res.Entries[idx].Counting = false
		throttle()
	}

	// 3) Endstand: nach Größe sortiert, gedeckelt.
	sort.Slice(res.Entries, func(i, j int) bool { return res.Entries[i].Size > res.Entries[j].Size })
	if len(res.Entries) > duMaxEntries {
		res.Entries = res.Entries[:duMaxEntries]
	}
	emitCopy()
}
