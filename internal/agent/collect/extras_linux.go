//go:build linux

package collect

import (
	"context"
	"strings"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

func osExtras(ctx context.Context) ([]shared.SoftwarePackage, []shared.Printer, []string) {
	return linuxSoftware(ctx), unixPrinters(ctx), unixUsers(ctx)
}

// linuxSoftware fragt das vorhandene Paketsystem ab (pacman/dpkg/rpm).
func linuxSoftware(ctx context.Context) []shared.SoftwarePackage {
	switch {
	case have("pacman"):
		return splitPkgs(run(ctx, "pacman", "-Q"), " ")
	case have("dpkg-query"):
		return splitPkgs(run(ctx, "dpkg-query", "-W", "-f", "${Package}\t${Version}\n"), "\t")
	case have("rpm"):
		return splitPkgs(run(ctx, "rpm", "-qa", "--qf", "%{NAME}\t%{VERSION}-%{RELEASE}\n"), "\t")
	}
	return nil
}

func splitPkgs(out, sep string) []shared.SoftwarePackage {
	var pkgs []shared.SoftwarePackage
	for _, line := range nonEmptyLines(out) {
		name, ver, _ := strings.Cut(line, sep)
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		pkgs = append(pkgs, shared.SoftwarePackage{Name: name, Version: strings.TrimSpace(ver)})
	}
	return pkgs
}
