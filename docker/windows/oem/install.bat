@echo off
REM Wird von dockur/windows nach der Windows-Installation einmalig als SYSTEM ausgefuehrt.
REM 1) DNS der VM auf den Samba-DC setzen (Voraussetzung fuer den Domaenenbeitritt)
REM 2) Verwaltungs-Features installieren (GPMC + AD-Benutzerverwaltung + DNS)
REM 3) Roster-Agent als Windows-Dienst installieren
REM Hinweis: agent.exe muss im OEM-Ordner liegen (per "make cross" erzeugt).

set "INSTALLDIR=%ProgramFiles%\Roster"
set "DATADIR=%ProgramData%\Roster"
set "DC_IP=172.19.0.10"
set "SERVER_IP=172.19.0.11"

REM --- DNS aller aktiven Adapter auf den Samba-DC zeigen lassen ---
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "Get-NetAdapter | Where-Object {$_.Status -eq 'Up'} | Set-DnsClientServerAddress -ServerAddresses '%DC_IP%'"
echo [Roster] DNS auf Samba-DC %DC_IP% gesetzt.

REM --- Verwaltungs-Tools: Gruppenrichtlinien (GPMC) + AD-Benutzer/-Computer (ADUC) + DNS ---
REM (GUI-Konsolen erfordern die "Desktop Experience"-Edition; auf Core nur das AD-PowerShell-Modul.)
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "Install-WindowsFeature -Name GPMC,RSAT-AD-Tools,RSAT-DNS-Server -IncludeManagementTools | Out-Null"
echo [Roster] Verwaltungs-Features installiert (GPMC, RSAT-AD-Tools, RSAT-DNS-Server).

REM --- Agent installieren (nur wenn vorhanden) ---
if not exist "%~dp0agent.exe" (
  echo [Roster] agent.exe fehlt im OEM-Ordner - Agent uebersprungen.
  goto :eof
)

mkdir "%INSTALLDIR%" 2>nul
mkdir "%DATADIR%" 2>nul
copy /Y "%~dp0agent.exe" "%INSTALLDIR%\agent.exe"

REM Vom Server erzeugtes TLS-Zertifikat aus dem geteilten Ordner pinnen.
copy /Y "\\172.30.0.1\Data\cert.pem" "%DATADIR%\cert.pem" >nul 2>&1

if not exist "%DATADIR%\agent.yaml" (
  > "%DATADIR%\agent.yaml" echo server_url: "https://%SERVER_IP%:8443"
  >> "%DATADIR%\agent.yaml" echo enrollment_token: "test-enroll-token"
  >> "%DATADIR%\agent.yaml" echo insecure_skip_verify: false
  >> "%DATADIR%\agent.yaml" echo ca_cert_path: "C:/ProgramData/Roster/cert.pem"
  >> "%DATADIR%\agent.yaml" echo interval: "30s"
  >> "%DATADIR%\agent.yaml" echo state_path: "C:/ProgramData/Roster/agent-state.json"
)

"%INSTALLDIR%\agent.exe" -config "%DATADIR%\agent.yaml" install
"%INSTALLDIR%\agent.exe" -config "%DATADIR%\agent.yaml" start
echo [Roster] Agent-Dienst installiert und gestartet.
