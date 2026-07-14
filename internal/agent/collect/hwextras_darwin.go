//go:build darwin

package collect

import (
	"context"
	"strings"

	"github.com/boonkerz/roster/internal/shared"
)

func hwExtras(ctx context.Context) ([]string, []shared.PhysicalDisk) {
	var gpus []string
	for _, l := range nonEmptyLines(run(ctx, "system_profiler", "SPDisplaysDataType")) {
		if i := strings.Index(l, "Chipset Model:"); i >= 0 {
			gpus = append(gpus, strings.TrimSpace(l[i+len("Chipset Model:"):]))
		}
	}
	return gpus, nil
}
