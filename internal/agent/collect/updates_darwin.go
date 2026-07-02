//go:build darwin

package collect

import (
	"context"
	"strings"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

// osUpdateList wertet `softwareupdate -l` aus (Anzahl/Namen verfügbarer Updates).
func osUpdateList(ctx context.Context) ([]shared.UpdateItem, bool) {
	out := run(ctx, "softwareupdate", "-l")
	if out == "" {
		return nil, true // kein Output -> i.d.R. keine Updates
	}
	var names []string
	for _, l := range nonEmptyLines(out) {
		if i := strings.Index(l, "Label:"); i >= 0 {
			names = append(names, strings.TrimSpace(l[i+len("Label:"):]))
		}
	}
	return packagesToItems(names), true
}
