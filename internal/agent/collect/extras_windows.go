//go:build windows

package collect

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

func osExtras(ctx context.Context) ([]shared.SoftwarePackage, []shared.Printer, []string) {
	return winSoftware(ctx), winPrinters(ctx), winUsers(ctx)
}

// psJSON führt ein PowerShell-Skript aus und parst dessen JSON-Ausgabe.
// ConvertTo-Json liefert bei einem Element ein Objekt, sonst ein Array – beides wird behandelt.
func psJSON[T any](ctx context.Context, script string) []T {
	out, err := psOutput(ctx, script)
	if err != nil {
		return nil
	}
	s := bytes.TrimSpace(out)
	if len(s) == 0 {
		return nil
	}
	var arr []T
	if err := json.Unmarshal(s, &arr); err == nil {
		return arr
	}
	var one T
	if err := json.Unmarshal(s, &one); err == nil {
		return []T{one}
	}
	return nil
}

func winSoftware(ctx context.Context) []shared.SoftwarePackage {
	const script = `$p=@('HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*',` +
		`'HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*'); ` +
		`Get-ItemProperty $p -ErrorAction SilentlyContinue | Where-Object {$_.DisplayName} | ` +
		`Select-Object @{n='Name';e={$_.DisplayName}},@{n='Version';e={$_.DisplayVersion}},` +
		`@{n='Publisher';e={$_.Publisher}} | ConvertTo-Json -Compress`
	type app struct{ Name, Version, Publisher string }
	var out []shared.SoftwarePackage
	for _, a := range psJSON[app](ctx, script) {
		if a.Name == "" {
			continue
		}
		out = append(out, shared.SoftwarePackage{Name: a.Name, Version: a.Version, Publisher: a.Publisher})
	}
	return out
}

func winPrinters(ctx context.Context) []shared.Printer {
	const script = `Get-CimInstance Win32_Printer | Select-Object Name,` +
		`@{n='Driver';e={$_.DriverName}},@{n='Port';e={$_.PortName}},Default | ConvertTo-Json -Compress`
	type printer struct {
		Name    string
		Driver  string
		Port    string
		Default bool
	}
	var out []shared.Printer
	for _, p := range psJSON[printer](ctx, script) {
		out = append(out, shared.Printer{Name: p.Name, Driver: p.Driver, Port: p.Port, Default: p.Default})
	}
	return out
}

func winUsers(ctx context.Context) []string {
	out, err := psOutput(ctx, "(Get-CimInstance Win32_ComputerSystem).UserName")
	if err != nil {
		return nil
	}
	if u := strings.TrimSpace(string(out)); u != "" {
		return []string{u}
	}
	return nil
}
