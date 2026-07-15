//go:build !windows

package main

import (
	"runtime"

	"github.com/ebitengine/purego"
)

// sdlLibHandle öffnet die bereits von SDL geladene Bibliothek erneut (dlopen ist
// referenzgezählt und liefert dasselbe Handle).
func sdlLibHandle() (uintptr, error) {
	names := []string{"libSDL3.so.0", "libSDL3.so"}
	if runtime.GOOS == "darwin" {
		names = []string{"libSDL3.dylib", "libSDL3.0.dylib"}
	}
	var err error
	for _, n := range names {
		var h uintptr
		if h, err = purego.Dlopen(n, purego.RTLD_LAZY); err == nil {
			return h, nil
		}
	}
	return 0, err
}
