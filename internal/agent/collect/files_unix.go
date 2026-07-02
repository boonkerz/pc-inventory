//go:build !windows

package collect

import (
	"encoding/json"
	"errors"
)

var errTooLarge = errors.New("Datei zu groß für den Transfer")

// rootsJSON liefert die Wurzel des Dateisystems ("/").
func rootsJSON() string {
	l := FileListing{Path: "", Entries: []FileEntry{{Name: "/", Path: "/", Dir: true}}}
	b, _ := json.Marshal(l)
	return string(b)
}
