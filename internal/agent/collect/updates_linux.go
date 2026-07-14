//go:build linux

package collect

import (
	"context"
	"os/exec"
	"strings"

	"github.com/boonkerz/roster/internal/shared"
)

// osUpdateList ermittelt aktualisierbare Pakete über das vorhandene Paketsystem.
// Es wird bewusst NICHT synchronisiert/aktualisiert (kein pacman -Sy / apt update),
// sondern nur der bereits bekannte Stand ausgewertet.
func osUpdateList(ctx context.Context) ([]shared.UpdateItem, bool) {
	var names []string
	switch {
	case have("checkupdates"): // pacman-contrib: nutzt eine temporäre DB, sicher
		names = firstFields(nonEmptyLines(run(ctx, "checkupdates")))
	case have("apt"):
		names = aptUpgradable(ctx)
	case have("dnf"):
		names = dnfUpdates(ctx)
	case have("apk"):
		names = apkUpgradable(ctx)
	case have("pacman"):
		names = firstFields(nonEmptyLines(run(ctx, "pacman", "-Qu")))
	default:
		return nil, false
	}
	return packagesToItems(names), true
}

// firstFields liefert je Zeile das erste Whitespace-Feld (den Paketnamen).
func firstFields(lines []string) []string {
	var out []string
	for _, l := range lines {
		if f := strings.Fields(l); len(f) > 0 {
			out = append(out, f[0])
		}
	}
	return out
}

func aptUpgradable(ctx context.Context) []string {
	var out []string
	for _, l := range nonEmptyLines(run(ctx, "apt", "list", "--upgradable")) {
		if !strings.Contains(l, "/") || strings.HasPrefix(l, "Listing") {
			continue
		}
		name, _, _ := strings.Cut(l, "/")
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

// apkUpgradable (Alpine) holt mit --no-cache einen frischen Index, ohne ihn zu
// persistieren, und listet aktualisierbare Pakete. Datenzeilen enthalten den
// "<"-Operator; Kopf- und "fetch"-Progresszeilen werden so herausgefiltert.
func apkUpgradable(ctx context.Context) []string {
	var out []string
	for _, l := range nonEmptyLines(run(ctx, "apk", "--no-cache", "version", "-l", "<")) {
		if !strings.Contains(l, "<") {
			continue
		}
		if f := strings.Fields(l); len(f) > 0 {
			out = append(out, f[0])
		}
	}
	return out
}

func dnfUpdates(ctx context.Context) []string {
	// check-update beendet sich mit Code 100, wenn Updates vorliegen -> stdout trotzdem nutzen.
	cmd := exec.CommandContext(ctx, "dnf", "-q", "--cacheonly", "check-update")
	b, _ := cmd.Output()
	var out []string
	for _, l := range nonEmptyLines(string(b)) {
		f := strings.Fields(l)
		if len(f) >= 3 && strings.Contains(f[0], ".") {
			out = append(out, f[0])
		}
	}
	return out
}
