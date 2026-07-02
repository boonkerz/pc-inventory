//go:build windows

package collect

import (
	"context"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

// osUpdateList fragt den Windows-Update-Agenten nach nicht installierten Updates inkl.
// Schweregrad und Support-URL. Der COM-Aufruf kann mehrere Sekunden dauern.
func osUpdateList(ctx context.Context) ([]shared.UpdateItem, bool) {
	const script = `$ProgressPreference='SilentlyContinue'; ` +
		`try { $s=New-Object -ComObject Microsoft.Update.Session; ` +
		`$r=$s.CreateUpdateSearcher().Search('IsInstalled=0 and IsHidden=0'); ` +
		`$r.Updates | ForEach-Object { [pscustomobject]@{Name=$_.Title; Severity=$_.MsrcSeverity; URL=$_.SupportUrl} } | ConvertTo-Json -Compress } catch { exit 1 }`
	type wu struct {
		Name     string
		Severity string
		URL      string
	}
	items := psJSON[wu](ctx, script)
	out := make([]shared.UpdateItem, 0, len(items))
	for _, it := range items {
		if it.Name == "" {
			continue
		}
		sev := it.Severity
		if sev == "" {
			sev = "Other"
		}
		out = append(out, shared.UpdateItem{Name: it.Name, Severity: sev, URL: it.URL})
	}
	return out, true
}
