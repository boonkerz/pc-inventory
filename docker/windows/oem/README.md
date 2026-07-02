# Windows-OEM-Ordner

Der Inhalt dieses Ordners wird von `dockur/windows` in die Windows-VM (nach `C:\OEM`)
kopiert; `install.bat` läuft **einmalig nach der Windows-Installation** als SYSTEM.

`install.bat` erledigt automatisch:
1. **DNS der VM auf den Samba-DC** (`172.19.0.10`) setzen – Voraussetzung für den Domänenbeitritt.
2. **Verwaltungs-Features** installieren: GPMC (Gruppenrichtlinien), RSAT-AD-Tools (AD-Benutzer
   und -Computer / ADUC) und RSAT-DNS-Server. (GUI-Konsolen brauchen die Desktop-Experience-Edition.)
3. Den **Agent** installieren und starten (Ziel: Inventory-Server `172.19.0.11:8443`).

Den Domänenbeitritt selbst stößt `join-domain.ps1` an (einmalig in der VM ausführen,
siehe unten) – bewusst getrennt, um den unbeaufsichtigten Erststart nicht durch einen
Neustart mitten in der Installation zu stören.

## Vorbereitung (vor `docker compose up`)

1. Windows-Agent cross-kompilieren:
   ```bash
   make cross   # erzeugt u.a. bin/agent-windows-amd64.exe
   ```
2. Binary hierher kopieren:
   ```bash
   cp bin/agent-windows-amd64.exe docker/windows/oem/agent.exe
   ```
3. Stack starten (Windows läuft automatisch mit):
   ```bash
   docker compose -f docker/compose.yml up --build
   ```

Fehlt `agent.exe`, überspringt `install.bat` nur den Agent-Teil (DNS wird trotzdem gesetzt).

## Domäne EXAMPLE.LOCAL beitreten

Nach dem Erststart in der VM (Konsole im Web-Viewer oder per RDP) einmalig:

```powershell
powershell -ExecutionPolicy Bypass -File C:\OEM\join-domain.ps1
```

Das Skript setzt DNS sicherheitshalber erneut auf den DC, prüft die SRV-Auflösung,
tritt mit `EXAMPLE\Administrator` / `Passw0rd!` bei und startet neu.

## Erststart

Die Windows-Installation läuft unbeaufsichtigt (deutsche Sprache/Tastatur); Fortschritt
unter http://localhost:8006. Der erste Lauf lädt das Eval-Image (mehrere GB) und dauert
einige Minuten. Windows startet erst, sobald der Samba-DC bereit ist.
