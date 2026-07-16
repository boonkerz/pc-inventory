package collect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// FileEntry ist ein Eintrag einer Verzeichnisauflistung.
type FileEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Dir      bool   `json:"dir"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"` // Unix-Sekunden
}

// FileListing ist das Ergebnis von BrowseDir.
type FileListing struct {
	Path    string      `json:"path"`
	Parent  string      `json:"parent"`
	Entries []FileEntry `json:"entries"`
	Error   string      `json:"error,omitempty"`
}

// MaxFileTransfer begrenzt Datei-Downloads/-Uploads (Schutz vor Speicher-/Zeitproblemen).
const MaxFileTransfer = 32 << 20 // 32 MiB

const maxListEntries = 2000

// BrowseDir listet die direkten Einträge eines Verzeichnisses (ohne Rekursion,
// schnell). Leerer Pfad -> Wurzeln (Laufwerke bzw. "/"). Liefert JSON.
func BrowseDir(path string) string {
	if path == "" {
		return rootsJSON()
	}
	l := FileListing{Path: path, Parent: filepath.Dir(path)}
	if l.Parent == path {
		l.Parent = ""
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		l.Error = err.Error()
		b, _ := json.Marshal(l)
		return string(b)
	}
	for _, e := range entries {
		info, ierr := e.Info()
		fe := FileEntry{Name: e.Name(), Path: filepath.Join(path, e.Name()), Dir: e.IsDir()}
		if ierr == nil {
			fe.Size = info.Size()
			fe.Modified = info.ModTime().Unix()
		}
		l.Entries = append(l.Entries, fe)
	}
	// Ordner zuerst, dann alphabetisch.
	sort.Slice(l.Entries, func(i, j int) bool {
		if l.Entries[i].Dir != l.Entries[j].Dir {
			return l.Entries[i].Dir
		}
		return l.Entries[i].Name < l.Entries[j].Name
	})
	if len(l.Entries) > maxListEntries {
		l.Entries = l.Entries[:maxListEntries]
	}
	b, _ := json.Marshal(l)
	return string(b)
}

// ReadFileCapped liest eine Datei bis MaxFileTransfer. Größere Dateien werden
// abgelehnt. Liefert den Inhalt oder einen Fehler.
func ReadFileCapped(path string) ([]byte, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, os.ErrInvalid
	}
	if fi.Size() > MaxFileTransfer {
		return nil, errTooLarge
	}
	return os.ReadFile(path)
}

// WriteFileCapped schreibt Daten in eine Datei (überschreibt, 0644).
func WriteFileCapped(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// MakeDir legt ein Verzeichnis an (inkl. fehlender Elternverzeichnisse).
func MakeDir(path string) error {
	if path == "" {
		return os.ErrInvalid
	}
	return os.MkdirAll(path, 0755)
}

// DeletePath entfernt eine Datei oder ein (leeres oder gefülltes) Verzeichnis.
func DeletePath(path string) error {
	if path == "" {
		return os.ErrInvalid
	}
	return os.RemoveAll(path)
}
