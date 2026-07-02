//go:build linux

package collect

import (
	"context"
	"os"
	"strings"
)

// hardwareIdentity liest Hersteller/Modell/Seriennummer aus dem DMI-Sysfs.
// Die Seriennummer erfordert meist root; ist sie nicht lesbar, bleibt sie leer.
func hardwareIdentity(_ context.Context) (vendor, model, serial string) {
	read := func(name string) string {
		b, err := os.ReadFile("/sys/class/dmi/id/" + name)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}
	return read("sys_vendor"), read("product_name"), read("product_serial")
}
