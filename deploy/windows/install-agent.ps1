<#
.SYNOPSIS
    Installiert den Roster-Agent als Windows-Dienst.

.DESCRIPTION
    Als Computer-Startskript per GPO verteilen. Das Skript ist idempotent:
    Bei bereits installiertem und enrolltem Agent passiert nichts.

    Verteilung typischerweise so:
      - agent-windows-amd64.exe + dieses Skript auf eine lesbare Netzwerkfreigabe legen
      - GPO -> Computerkonfiguration -> Richtlinien -> Windows-Einstellungen
        -> Skripts -> Startup -> dieses Skript mit Parametern eintragen.

.PARAMETER ServerUrl
    URL des Inventory-Servers, z.B. https://inventory.example.com:8443

.PARAMETER EnrollmentToken
    In der Web-UI erzeugtes Enrollment-Token.

.PARAMETER SourceExe
    Pfad zur agent-windows-amd64.exe (z.B. auf der Freigabe).

.PARAMETER CaCertSource
    Optional: Pfad zum CA-Zertifikat des Servers (wird mitkopiert und gepinnt).
#>
param(
    [Parameter(Mandatory = $true)] [string] $ServerUrl,
    [Parameter(Mandatory = $true)] [string] $EnrollmentToken,
    [Parameter(Mandatory = $true)] [string] $SourceExe,
    [string] $CaCertSource = ""
)

$ErrorActionPreference = "Stop"
$ServiceName = "roster-agent"
$InstallDir  = Join-Path $env:ProgramFiles "Roster"
$DataDir     = Join-Path $env:ProgramData "Roster"
$ExePath     = Join-Path $InstallDir "agent.exe"
$ConfigPath  = Join-Path $DataDir "agent.yaml"

New-Item -ItemType Directory -Force -Path $InstallDir, $DataDir | Out-Null

# Binary kopieren bzw. aktualisieren.
Copy-Item -Path $SourceExe -Destination $ExePath -Force

$CaLine = "ca_cert_path: `"`""
if ($CaCertSource -ne "") {
    $CaTarget = Join-Path $DataDir "server-ca.crt"
    Copy-Item -Path $CaCertSource -Destination $CaTarget -Force
    $CaLine = "ca_cert_path: `"$($CaTarget -replace '\\','\\')`""
}

# Konfiguration nur schreiben, wenn sie noch nicht existiert,
# damit ein bereits enrollter Agent sein Token behält.
if (-not (Test-Path $ConfigPath)) {
@"
server_url: "$ServerUrl"
enrollment_token: "$EnrollmentToken"
$CaLine
insecure_skip_verify: false
interval: "5m"
state_path: "$($DataDir -replace '\\','\\')\\agent-state.json"
"@ | Set-Content -Path $ConfigPath -Encoding UTF8
}

# Dienst installieren (idempotent) und starten.
$existing = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if (-not $existing) {
    & $ExePath -config $ConfigPath install
}
& $ExePath -config $ConfigPath start

Write-Host "Roster-Agent installiert und gestartet."
