//go:build windows

package collect

import (
	"context"
	"strings"
)

// hardwareIdentity fragt Hersteller/Modell/Seriennummer über CIM/WMI per PowerShell ab.
// (Eine native WMI-Anbindung kann später folgen; PowerShell ist überall vorhanden.)
func hardwareIdentity(ctx context.Context) (vendor, model, serial string) {
	vendor = cimQuery(ctx, "Win32_ComputerSystem", "Manufacturer")
	model = cimQuery(ctx, "Win32_ComputerSystem", "Model")
	serial = cimQuery(ctx, "Win32_BIOS", "SerialNumber")
	return vendor, model, serial
}

func cimQuery(ctx context.Context, class, property string) string {
	ps := "(Get-CimInstance -ClassName " + class + ")." + property
	out, err := psOutput(ctx, ps)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
