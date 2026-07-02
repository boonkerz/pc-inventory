//go:build !windows

package update

import (
	"os"
	"syscall"
)

// restart ersetzt den laufenden Prozess durch das (soeben aktualisierte) Binary
// unter exe. Gleiche PID – unter systemd/launchd ist kein Dienst-Neustart nötig.
func restart(exe string) error {
	return syscall.Exec(exe, os.Args, os.Environ())
}
