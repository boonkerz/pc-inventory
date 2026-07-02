package collect

import (
	"context"
	"strings"
	"testing"
)

func TestListProcessesSmoke(t *testing.T) {
	out := ListProcesses(context.Background())
	if !strings.Contains(out, `"processes"`) {
		t.Fatalf("keine Prozessliste: %s", out[:min(200, len(out))])
	}
}

func TestListServicesSmoke(t *testing.T) {
	out := ListServices(context.Background())
	// Auf Linux sollte "services" da sein; enthält sonst evtl. "error" (kein systemd).
	if !strings.Contains(out, `"services"`) && !strings.Contains(out, `"error"`) {
		t.Fatalf("unerwartete Dienstausgabe: %s", out[:min(200, len(out))])
	}
	t.Logf("services sample: %.200s", out)
}

func min(a, b int) int { if a < b { return a }; return b }
