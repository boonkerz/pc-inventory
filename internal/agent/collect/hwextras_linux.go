//go:build linux

package collect

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/boonkerz/roster/internal/shared"
)

func hwExtras(ctx context.Context) ([]string, []shared.PhysicalDisk) {
	return linuxGPUs(ctx), linuxPhysicalDisks()
}

func linuxGPUs(ctx context.Context) []string {
	if !have("lspci") {
		return nil
	}
	var out []string
	for _, l := range nonEmptyLines(run(ctx, "lspci")) {
		if strings.Contains(l, "VGA compatible controller") ||
			strings.Contains(l, "3D controller") ||
			strings.Contains(l, "Display controller") {
			if i := strings.LastIndex(l, ": "); i >= 0 {
				out = append(out, strings.TrimSpace(l[i+2:]))
			}
		}
	}
	return out
}

// linuxPhysicalDisks liest physische Laufwerke aus dem Sysfs (/sys/block).
func linuxPhysicalDisks() []shared.PhysicalDisk {
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil
	}
	var out []shared.PhysicalDisk
	for _, e := range entries {
		name := e.Name()
		if hasAnyPrefix(name, "loop", "ram", "dm-", "sr", "zram", "md", "fd") {
			continue
		}
		sectors, _ := strconv.ParseUint(strings.TrimSpace(readSysfs("/sys/block/"+name+"/size")), 10, 64)
		size := sectors * 512
		if size == 0 {
			continue
		}
		model := strings.TrimSpace(readSysfs("/sys/block/" + name + "/device/model"))
		if model == "" {
			model = name
		}
		out = append(out, shared.PhysicalDisk{Model: model, SizeBytes: size})
	}
	return out
}

func readSysfs(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
