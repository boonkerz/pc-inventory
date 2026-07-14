# Test-Stack (Docker Compose)

Eine in sich geschlossene Testumgebung für Roster mit vier Komponenten:

| Service           | Rolle                                   | Status / Hinweis                          |
|-------------------|-----------------------------------------|-------------------------------------------|
| `inventory-host`  | Inventar-Server (Go, SQLite, ohne TLS)  | Web-UI auf http://localhost:8443          |
| `linux-test-pc`   | Linux-Client, lässt nur den Agent laufen| meldet sich automatisch per Seed-Token an |
| `samba-addc`      | Samba Active Directory Domain Controller| standalone, bereit für LDAP/M4            |
| `windows-test-pc` | Windows Server (headless) als KVM-VM    | startet mit dem Stack (braucht /dev/kvm)  |

## Schnellstart

```bash
# Windows-Agent vorab bereitstellen (sonst läuft der Windows-PC ohne Agent):
make cross && cp bin/agent-windows-amd64.exe docker/windows/oem/agent.exe

docker compose -f docker/compose.yml up --build   # startet ALLES inkl. Windows-VM
```

Danach:

- **Web-UI**: http://localhost:8443 — Login `admin` / `admin1234`
- Der Linux-Agent registriert sich automatisch (festes `ROSTER_SEED_ENROLL_TOKEN=test-enroll-token`,
  nur für diese Testumgebung) und erscheint als Gerät `linux-test-pc` mit Status *online*.
- Der **Windows-PC** wird unbeaufsichtigt (deutsch) installiert; Fortschritt unter
  http://localhost:8006. Der Erststart lädt mehrere GB und dauert einige Minuten.

> Kein KVM verfügbar oder Windows nicht gewünscht? Den `windows-test-pc`-Service in
> `compose.yml` auskommentieren oder gezielt nur die Linux-Dienste starten:
> `docker compose -f docker/compose.yml up inventory-host linux-test-pc samba-addc`.

## Samba Active Directory

Beim ersten Start wird die Domäne **EXAMPLE.LOCAL** provisioniert (dauert ~20–30 s):

- Admin: `Administrator` / `Passw0rd!`
- Testbenutzer: `tester` / `Test1234`, `viewer` / `View1234`
- Gruppe: `inventory-admins` (Mitglied: tester)
- LDAP `localhost:1389`, LDAPS `localhost:1636` (auf den Host veröffentlicht)

Wichtig für das spätere LDAP-Login (M4): Active Directory lehnt Simple-Binds über
**unverschlüsseltes** LDAP ab („Transport encryption required"). Der Bind muss über
**LDAPS** (Port 636) bzw. StartTLS erfolgen — verifiziert mit:

```bash
docker compose -f docker/compose.yml exec -e LDAPTLS_REQCERT=never samba-addc \
  ldapsearch -H ldaps://localhost:636 -x -D "tester@EXAMPLE.LOCAL" -w Test1234 \
  -b "DC=example,DC=local" "(sAMAccountName=tester)"
```

## Feste IPs

Damit DNS- und Agent-Ziele stabil sind, nutzt `invnet` ein festes `/24`-Subnetz:

| Adresse        | Service          | Zweck                                  |
|----------------|------------------|----------------------------------------|
| `172.19.0.10`  | `samba-addc`     | DNS-Server der Windows-VM (Domänen-DNS) |
| `172.19.0.11`  | `inventory-host` | Ziel des Windows-Agents (per IP)        |

## Windows-Test-PC

Echte Windows-Container brauchen einen Windows-Host. Unter Linux läuft Windows nur als
KVM-VM (`dockur/windows`) – der Service startet zusammen mit dem übrigen Stack. Damit der
Agent in der VM installiert wird, vorab das Binary bereitstellen (siehe `windows/oem/README.md`):

```bash
make cross                                   # erzeugt bin/agent-windows-amd64.exe
cp bin/agent-windows-amd64.exe docker/windows/oem/agent.exe
```

- **Deutsch**: Sprache, Region und Tastatur werden per `LANGUAGE`/`REGION`/`KEYBOARD`
  auf Deutsch (`de-DE`) installiert. Das greift nur bei einer **frischen** Installation.
- Windows startet erst, wenn der Samba-DC *healthy* ist (`depends_on`), damit DNS bereitsteht.
- Nach der unbeaufsichtigten Installation erledigt `oem/install.bat` automatisch:
  **DNS auf den Samba-DC** (172.19.0.10), **Verwaltungs-Features** (GPMC, RSAT-AD-Tools,
  RSAT-DNS-Server) und die **Agent-Installation** (Ziel `172.19.0.11`).
- Installationsfortschritt: http://localhost:8006 (Erststart lädt mehrere GB, dauert).

> **Verwaltungs-Tools**: GPMC (Gruppenrichtlinienverwaltung) und ADUC (AD-Benutzer und
> -Computer) sind GUI-Konsolen und brauchen die **Desktop-Experience**-Edition. Auf einer
> bestehenden VM nachinstallieren (PowerShell als Admin):
> ```powershell
> Install-WindowsFeature -Name GPMC,RSAT-AD-Tools,RSAT-DNS-Server -IncludeManagementTools
> ```
> Aufruf danach: `gpmc.msc` bzw. `dsa.msc`.

### Agent in die Windows-VM bringen

- **Frische Installation**: `agent.exe` im `oem/`-Ordner genügt – `install.bat` installiert
  den Dienst automatisch (siehe oben).
- **Bestehende VM** (ohne Neuinstallation): die `agent.exe` über die Freigabe **`\\172.30.0.1\Data`**
  hineinkopieren (IP statt `host.lan`, da die VM den Samba-DC als DNS nutzt und den dockur-Namen
  nicht auflöst; `172.30.0.1` ist das interne Gateway, auf dem `smbd` lauscht). Datei dorthin legen mit
  `docker cp docker/windows/oem/agent.exe roster-test-windows-test-pc-1:/tmp/smb/`
  (oder dauerhaft via `docker/windows/shared/`), dann in der VM (PowerShell als Admin):
  ```powershell
  $dst = "$env:ProgramFiles\Roster"
  New-Item -ItemType Directory -Force $dst, "$env:ProgramData\Roster" | Out-Null
  Copy-Item \\172.30.0.1\Data\agent.exe "$dst\agent.exe"
  @"
  server_url: "http://172.19.0.11:8443"
  enrollment_token: "test-enroll-token"
  insecure_skip_verify: true
  interval: "30s"
  state_path: "C:/ProgramData/Roster/agent-state.json"
  "@ | Set-Content "$env:ProgramData\Roster\agent.yaml" -Encoding ascii
  & "$dst\agent.exe" -config "$env:ProgramData\Roster\agent.yaml" install
  & "$dst\agent.exe" -config "$env:ProgramData\Roster\agent.yaml" start
  ```

### Der Domäne EXAMPLE.LOCAL beitreten

DNS zeigt nach `install.bat` bereits auf den DC. In der VM (Konsole im Web-Viewer oder per
RDP) einmalig ausführen:

```powershell
powershell -ExecutionPolicy Bypass -File C:\OEM\join-domain.ps1
```

Das Skript prüft die SRV-Auflösung, tritt mit `EXAMPLE\Administrator` / `Passw0rd!` bei und
startet neu. Danach Anmeldung mit Domänen-Benutzern (z. B. `EXAMPLE\tester` / `Test1234`).

> **Hinweis:** Die festen IPs, das `/24`-Subnetz und das deutsche Windows greifen nur bei
> einem Neuaufbau. Bestehenden Stack dafür zurücksetzen:
> `docker compose -f docker/compose.yml down -v` (löscht auch Domäne + Windows-Installation),
> danach neu `up`.

## Aufräumen

```bash
docker compose -f docker/compose.yml down            # Container stoppen (Volumes bleiben)
docker compose -f docker/compose.yml down -v         # inkl. Daten/Domäne zurücksetzen
```
