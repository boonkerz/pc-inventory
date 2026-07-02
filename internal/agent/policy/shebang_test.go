package policy

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestRunScriptHonorsBashShebang(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("nur unix")
	}
	// Nutzt bash-only-Syntax ([[ ]], set -o pipefail), die unter dash (sh) scheitern würde.
	script := "#!/usr/bin/env bash\nset -euo pipefail\nx=5\nif [[ \"$x\" == \"5\" ]]; then echo ok-bash; fi\n"
	exit, out, ok := RunScript(context.Background(), "shell", script, nil)
	if !ok {
		t.Fatal("sollte anwendbar sein")
	}
	if exit != 0 || !strings.Contains(out, "ok-bash") {
		t.Fatalf("bash-Skript scheiterte: exit=%d out=%q", exit, out)
	}
}

func TestRunScriptPlainShellStillWorks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("nur unix")
	}
	// Ohne Shebang -> weiterhin sh -c
	exit, out, ok := RunScript(context.Background(), "shell", "echo hallo; exit 0", nil)
	if !ok || exit != 0 || !strings.Contains(out, "hallo") {
		t.Fatalf("plain shell: exit=%d out=%q ok=%v", exit, out, ok)
	}
}
