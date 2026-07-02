// Package web bettet das gebaute React-Frontend (web/dist) in das Server-Binary ein.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS liefert das Wurzelverzeichnis des gebauten Frontends (dist/).
// Gibt nil zurück, falls noch kein Build vorhanden ist.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil
	}
	return sub
}
