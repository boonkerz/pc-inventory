//go:build darwin

package collect

import (
	"context"
	"os"
	"strings"

	"github.com/boonkerz/roster/internal/shared"
)

func osExtras(ctx context.Context) ([]shared.SoftwarePackage, []shared.Printer, []string) {
	return macSoftware(), unixPrinters(ctx), unixUsers(ctx)
}

// macSoftware listet installierte Anwendungen aus den App-Verzeichnissen.
func macSoftware() []shared.SoftwarePackage {
	var apps []shared.SoftwarePackage
	for _, dir := range []string{"/Applications", "/System/Applications"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".app") {
				apps = append(apps, shared.SoftwarePackage{Name: strings.TrimSuffix(e.Name(), ".app")})
			}
		}
	}
	return apps
}
