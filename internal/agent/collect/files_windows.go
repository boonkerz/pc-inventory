//go:build windows

package collect

import (
	"encoding/json"
	"errors"
	"os"
)

var errTooLarge = errors.New("Datei zu groß für den Transfer")

// rootsJSON liefert die verfügbaren Laufwerksbuchstaben.
func rootsJSON() string {
	l := FileListing{Path: ""}
	for c := 'A'; c <= 'Z'; c++ {
		p := string(c) + ":\\"
		if _, err := os.Stat(p); err == nil {
			l.Entries = append(l.Entries, FileEntry{Name: p, Path: p, Dir: true})
		}
	}
	b, _ := json.Marshal(l)
	return string(b)
}
