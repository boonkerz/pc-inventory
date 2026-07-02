//go:build windows

package collect

import (
	"context"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

func hwExtras(ctx context.Context) ([]string, []shared.PhysicalDisk) {
	return winGPUs(ctx), winPhysicalDisks(ctx)
}

func winGPUs(ctx context.Context) []string {
	type vc struct{ Name string }
	var out []string
	for _, g := range psJSON[vc](ctx, `Get-CimInstance Win32_VideoController | Select-Object Name | ConvertTo-Json -Compress`) {
		if g.Name != "" {
			out = append(out, g.Name)
		}
	}
	return out
}

func winPhysicalDisks(ctx context.Context) []shared.PhysicalDisk {
	type dd struct {
		Model string
		Size  float64 // Win32_DiskDrive.Size kommt als (großer) JSON-Zahlenwert
	}
	var out []shared.PhysicalDisk
	for _, d := range psJSON[dd](ctx, `Get-CimInstance Win32_DiskDrive | Select-Object Model,Size | ConvertTo-Json -Compress`) {
		out = append(out, shared.PhysicalDisk{Model: d.Model, SizeBytes: uint64(d.Size)})
	}
	return out
}
