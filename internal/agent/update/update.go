// Package update implementiert das Selbst-Update des Agents: das passende Binary
// wird vom Inventory-Server geladen, ersetzt das laufende Binary und der Agent
// startet neu.
package update

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/minio/selfupdate"

	"github.com/boonkerz/roster/internal/agent/transport"
)

// Platform liefert den Plattform-Schlüssel ("<os>-<arch>") für den Download.
func Platform() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

// ShouldUpdate entscheidet, ob auf latest aktualisiert werden soll. Dev-Builds und
// leere/gleiche Versionen werden übersprungen, um Update-Schleifen zu vermeiden.
func ShouldUpdate(current, latest string) bool {
	if latest == "" || latest == "dev" || current == "dev" {
		return false
	}
	return latest != current
}

// Apply lädt das Binary für die eigene Plattform, ersetzt das laufende Binary und
// startet den Agent neu. Auf Unix kehrt die Funktion bei Erfolg nicht zurück (re-exec).
func Apply(ctx context.Context, client *transport.Client) error {
	// Pfad VOR dem Ersetzen erfassen: danach liefert os.Executable() auf Linux den
	// Pfad mit "(deleted)"-Suffix, weil selfupdate das alte Inode ersetzt.
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("eigenen pfad ermitteln: %w", err)
	}

	rc, err := client.DownloadAgent(ctx, Platform())
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer rc.Close()

	if err := selfupdate.Apply(rc, selfupdate.Options{}); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("update fehlgeschlagen, rollback ebenfalls fehlgeschlagen: %w", rerr)
		}
		return fmt.Errorf("update anwenden: %w", err)
	}
	return restart(exe)
}
