//go:build windows

package collect

import (
	"context"
	"encoding/json"
	"strings"
)

// AVStatusJSON liest den Defender-Status via Get-MpComputerStatus.
func AVStatusJSON(ctx context.Context) string {
	out, err := psOutput(ctx, `Get-MpComputerStatus | Select-Object @{N='p';E={'Windows Defender'}},@{N='en';E={$_.AntivirusEnabled}},@{N='rt';E={$_.RealTimeProtectionEnabled}},@{N='age';E={[int]$_.AntivirusSignatureAge}},@{N='ver';E={"$($_.AMProductVersion)"}} | ConvertTo-Json -Compress`)
	if err != nil {
		return jsonOrError(nil, err)
	}
	var r struct {
		P   string `json:"p"`
		En  bool   `json:"en"`
		Rt  bool   `json:"rt"`
		Age int    `json:"age"`
		Ver string `json:"ver"`
	}
	if json.Unmarshal(out, &r) != nil {
		return jsonOrError(nil, errUnparsable)
	}
	return jsonOrError(AVStatus{Product: r.P, Enabled: r.En, RealTime: r.Rt, SignatureAge: r.Age, Version: r.Ver}, nil)
}

// BitLockerJSON liest je Volume Schutzstatus + Wiederherstellungsschlüssel.
func BitLockerJSON(ctx context.Context) string {
	script := `Get-BitLockerVolume | ForEach-Object {
  $rp = ($_.KeyProtector | Where-Object { $_.KeyProtectorType -eq 'RecoveryPassword' } | Select-Object -First 1)
  [PSCustomObject]@{ mp="$($_.MountPoint)"; prot="$($_.ProtectionStatus)"; pct=[int]$_.EncryptionPercentage; key="$($rp.RecoveryPassword)"; rid="$($rp.KeyProtectorId)" }
} | ConvertTo-Json -Compress`
	out, err := psOutput(ctx, script)
	if err != nil {
		return jsonOrError(nil, err)
	}
	type row struct {
		Mp   string `json:"mp"`
		Prot string `json:"prot"`
		Pct  int    `json:"pct"`
		Key  string `json:"key"`
		Rid  string `json:"rid"`
	}
	var rows []row
	trimmed := strings.TrimSpace(string(out))
	if strings.HasPrefix(trimmed, "[") {
		_ = json.Unmarshal(out, &rows)
	} else if trimmed != "" {
		var one row
		if json.Unmarshal(out, &one) == nil {
			rows = []row{one}
		}
	}
	vols := make([]BitLockerVolume, 0, len(rows))
	for _, r := range rows {
		prot := "Unknown"
		switch r.Prot {
		case "On", "1":
			prot = "On"
		case "Off", "0":
			prot = "Off"
		}
		vols = append(vols, BitLockerVolume{MountPoint: r.Mp, Protection: prot, Percent: r.Pct, RecoveryKey: r.Key, RecoveryID: r.Rid})
	}
	return jsonOrError(map[string]any{"volumes": vols}, nil)
}

// SmartJSON liest die Datenträgergesundheit via Get-PhysicalDisk.
func SmartJSON(ctx context.Context) string {
	out, err := psOutput(ctx, `Get-PhysicalDisk | Select-Object @{N='n';E={"$($_.DeviceId)"}},@{N='m';E={"$($_.FriendlyName)"}},@{N='h';E={"$($_.HealthStatus)"}} | ConvertTo-Json -Compress`)
	if err != nil {
		return jsonOrError(nil, err)
	}
	type row struct{ N, M, H string }
	var rows []row
	trimmed := strings.TrimSpace(string(out))
	if strings.HasPrefix(trimmed, "[") {
		_ = json.Unmarshal(out, &rows)
	} else if trimmed != "" {
		var one row
		if json.Unmarshal(out, &one) == nil {
			rows = []row{one}
		}
	}
	disks := make([]SmartDisk, 0, len(rows))
	for _, r := range rows {
		h := "Unbekannt"
		switch r.H {
		case "Healthy":
			h = "OK"
		case "Warning":
			h = "Warnung"
		case "Unhealthy":
			h = "Fehler"
		}
		disks = append(disks, SmartDisk{Name: r.N, Model: r.M, Health: h, Detail: r.H})
	}
	return jsonOrError(map[string]any{"disks": disks}, nil)
}

// EventLogJSON liest die letzten Ereignisse eines Windows-Logs (System|Application).
func EventLogJSON(ctx context.Context, logName string, count int) string {
	if logName != "System" && logName != "Application" && logName != "Security" {
		logName = "System"
	}
	if count <= 0 || count > 200 {
		count = 100
	}
	script := `Get-WinEvent -LogName ` + logName + ` -MaxEvents ` + itoa(count) +
		` | Select-Object @{N='t';E={$_.TimeCreated.ToString('o')}},@{N='id';E={$_.Id}},@{N='lvl';E={"$($_.LevelDisplayName)"}},@{N='src';E={"$($_.ProviderName)"}},@{N='msg';E={"$($_.Message)".Substring(0,[Math]::Min(300,"$($_.Message)".Length))}} | ConvertTo-Json -Compress`
	out, err := psOutput(ctx, script)
	if err != nil {
		return jsonOrError(nil, err)
	}
	return wrapEvents(out)
}
