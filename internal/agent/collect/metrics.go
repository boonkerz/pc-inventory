package collect

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// Metrics ist eine Momentaufnahme der Systemauslastung.
type Metrics struct {
	Timestamp       int64      `json:"timestamp"` // Unix-Millisekunden
	CPUPercent      float64    `json:"cpu_percent"`
	CPUPerCore      []float64  `json:"cpu_per_core,omitempty"`
	LoadAvg         [3]float64 `json:"load_avg,omitempty"`
	MemUsedPercent  float64    `json:"mem_used_percent"`
	MemUsed         uint64     `json:"mem_used"`
	MemTotal        uint64     `json:"mem_total"`
	SwapUsedPercent float64    `json:"swap_used_percent"`
	Disks           []DiskUse  `json:"disks,omitempty"`
	Net             []NetIO    `json:"net,omitempty"`
}

// DiskUse ist die Belegung eines Mountpoints.
type DiskUse struct {
	Name        string  `json:"name"`
	UsedPercent float64 `json:"used_percent"`
	Used        uint64  `json:"used"`
	Total       uint64  `json:"total"`
}

// NetIO sind die kumulativen Byte-Zähler einer Schnittstelle (Client bildet die Rate).
type NetIO struct {
	Name      string `json:"name"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
}

// MetricsJSON sammelt eine Momentaufnahme und liefert sie als JSON. Die CPU-Messung
// blockiert kurz (ein Messintervall), daher läuft der Befehl asynchron.
func MetricsJSON(ctx context.Context) string {
	m := Metrics{Timestamp: time.Now().UnixMilli()}

	if pct, err := cpu.PercentWithContext(ctx, 500*time.Millisecond, false); err == nil && len(pct) > 0 {
		m.CPUPercent = pct[0]
	}
	if per, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
		m.CPUPerCore = per
	}
	if l, err := load.AvgWithContext(ctx); err == nil {
		m.LoadAvg = [3]float64{l.Load1, l.Load5, l.Load15}
	}
	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		m.MemUsedPercent = vm.UsedPercent
		m.MemUsed = vm.Used
		m.MemTotal = vm.Total
	}
	if sw, err := mem.SwapMemoryWithContext(ctx); err == nil {
		m.SwapUsedPercent = sw.UsedPercent
	}
	if parts, err := disk.PartitionsWithContext(ctx, false); err == nil {
		seen := map[string]bool{}
		for _, p := range parts {
			if !realFSTypes[strings.ToLower(p.Fstype)] || seen[p.Device] {
				continue
			}
			u, err := disk.UsageWithContext(ctx, p.Mountpoint)
			if err != nil || u.Total == 0 {
				continue
			}
			seen[p.Device] = true
			m.Disks = append(m.Disks, DiskUse{Name: p.Mountpoint, UsedPercent: u.UsedPercent, Used: u.Used, Total: u.Total})
		}
	}
	if io, err := net.IOCountersWithContext(ctx, true); err == nil {
		for _, n := range io {
			if n.BytesSent == 0 && n.BytesRecv == 0 {
				continue
			}
			m.Net = append(m.Net, NetIO{Name: n.Name, BytesSent: n.BytesSent, BytesRecv: n.BytesRecv})
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}
