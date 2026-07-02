package collect

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/shirou/gopsutil/v4/process"
)

// ProcInfo beschreibt einen laufenden Prozess (nach Speicher sortiert).
type ProcInfo struct {
	PID      int32  `json:"pid"`
	Name     string `json:"name"`
	User     string `json:"user,omitempty"`
	MemBytes uint64 `json:"mem_bytes"`
}

const maxProcs = 400

// ListProcesses liefert die laufenden Prozesse als JSON (nach RSS absteigend).
func ListProcesses(ctx context.Context) string {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(b)
	}
	out := make([]ProcInfo, 0, len(procs))
	for _, p := range procs {
		if ctx.Err() != nil {
			break
		}
		name, _ := p.NameWithContext(ctx)
		if name == "" {
			continue
		}
		var mem uint64
		if mi, e := p.MemoryInfoWithContext(ctx); e == nil && mi != nil {
			mem = mi.RSS
		}
		user, _ := p.UsernameWithContext(ctx)
		out = append(out, ProcInfo{PID: p.Pid, Name: name, User: user, MemBytes: mem})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MemBytes > out[j].MemBytes })
	if len(out) > maxProcs {
		out = out[:maxProcs]
	}
	b, _ := json.Marshal(map[string]any{"processes": out})
	return string(b)
}

// KillProcess beendet einen Prozess anhand seiner PID.
func KillProcess(pid int32) (int, string) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 1, "Prozess nicht gefunden"
	}
	if err := p.Kill(); err != nil {
		return 1, err.Error()
	}
	return 0, "Prozess beendet"
}
