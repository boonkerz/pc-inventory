# PC-Inventory – Feature-Liste / Roadmap

Stand: laufend gepflegt. Reihenfolge unter „Als nächstes" = aktuelle Priorität.

## Als nächstes (in Arbeit)

- [x] **Software-Änderungs-Tracking** – Software-Inventar zwischen Checkins diffen,
  neu installierte / entfernte / aktualisierte Programme protokollieren (Verlauf).
  Offen: optional alarmieren.
- [x] **Übersichts-Dashboard** – Health-Zusammenfassung über alle Geräte: fehlschlagende
  Checks, ausstehende Patches, Offline-Geräte, Task-Fehler; Status-Donut + letzte Wechsel.

## Geplant – schnelle Gewinne (nutzen vorhandene Plumbing)

- [x] **Wartungsfenster / Alarme stummschalten** – Zeitfenster pro Client/Site/Gerät, in dem
  `alertTransitions` nichts sendet (Checks laufen + Verlauf bleibt). Verwaltung in Einstellungen.
- [x] **Native Netzwerk-Checks** – Ping / TCP-Port / HTTP-Status als eigene Check-Typen
  (mit optionalen Latenz-Schwellen; Platzhalter in Host/URL nutzbar).
- [x] **Dienste & Prozesse** – Windows-Dienste / systemd-Units + Prozesse on-demand,
  mit Start/Stop/Neustart bzw. Beenden über die Command-Queue (Tab „Dienste/Prozesse").
- [x] **Wake-on-LAN** – Magic Packet von einem online Nachbar-Agent im selben Standort
  („Aufwecken"-Button bei offline Geräten).
- [x] **Sammelaktionen (Bulk)** – Skript ausführen oder Update-Scan auf allen Geräten
  eines Ziels (Client/Site/Tag/alle); Nav „Sammelaktion".

## Geplant – größer, hoher Wert

- [x] **Audit-Log** – wer hat wann was getan (Login/Fehlversuche + alle ändernden
  Aktionen via Middleware); Ansicht in Einstellungen.
- [x] **Feingranulare Rollen (RBAC)** – Viewer (nur lesen), Techniker (Geräte bedienen:
  Skripte/Dienste/Prozesse/Updates/Neustart/WoL/Notizen/Sammelaktion), Admin (alles).
- [x] **Dateibrowser / -transfer** – Verzeichnisse durchsuchen, Dateien
  herunterladen/hochladen (bis 32 MB) über dedizierte Transfer-Endpoints. Tab „Dateien".
- [x] **Geplante Reports (E-Mail/HTML)** – Health-Bericht je Kunde (Geräte online/offline,
  fehlerhafte Checks/Tasks, ausstehende Patches). On-demand als HTML (druckbar zu PDF)
  + geplanter Versand (täglich/wöchentlich/monatlich) über einen Alarm-Kanal.

## Sicherheits-/Inventar-Collectors (Tabs „Sicherheit" + „Ereignisse")

- [x] **Defender/AV-Status** (Windows Get-MpComputerStatus; Linux ClamAV-Hinweis).
- [x] **BitLocker-Status** (+ Recovery-Key-Escrow serverseitig).
- [x] **SMART-Festplattengesundheit** (Windows Get-PhysicalDisk, Linux smartctl).
- [x] **Windows-Event-Log / journald-Viewer** (on-demand, Filter).
- [x] **Geräte-Notizen / Doku** pro Gerät (Übersicht-Tab, auch durchsuchbar).

## Erledigt (Auszug)

- [x] Inventar (Hardware, Software, Netzwerk, Datenträger), cross-platform Agents mit Auto-Update
- [x] Checks (Disk/Memory/CPU/Updates/Script) mit Frequenz, Schweregrad, Ausgabe-Vergleich, Plattform-Targeting
- [x] Tasks (geplante Skripte) mit Frequenz; letzter Lauf je Task + Lauf-Historie
- [x] Custom Fields (TRMM-Stil) + JSON-Collector + Twig-artige Platzhalter mit Filtern
- [x] Modulares Alerting (E-Mail/Webhook/Pushover/Telegram/ntfy), Scope + Schweregrad, Recovery-Meldung
- [x] Check-Statuswechsel-Verlauf (Historie) inkl. Benachrichtigungs-Status
- [x] Remote-Terminal (On-demand Wake-Poll) + Popout-Fenster
- [x] Patch-Management (Scan/Genehmigen/Installieren)
- [x] Ad-hoc-Skripte mit Push-Trigger; Command-Queue
- [x] 2FA (TOTP, Pflicht) + Backup-Codes
- [x] Organisation: Client/Site/Device-Hierarchie, Tags/Gruppen
- [x] Neustart, Token-Widerruf
- [x] TreeSize-Speicheransicht (live hochzählend, Tortendiagramm)
- [x] Konto-Selbstverwaltung (Web-UI): Kontodaten + Passwort ändern
- [x] Konto-Wiederherstellung per CLI (list-users, reset-password, disable-2fa)
- [x] Historie-Pruning (30 Tage), Deploy hinter Reverse-Proxy
- [x] Dienste & Prozesse (on-demand + Steuerung), Wake-on-LAN, Sammelaktionen
- [x] Globale Gerätesuche (Hostname, IP/MAC, OS, Seriennr., Software, Custom Fields)

## Offene Kleinigkeiten / Schulden

- [x] **2FA-Reset durch Admin** in der Web-UI (Benutzerliste → „2FA zurücksetzen").
- [x] **Software-Änderungen alarmieren** (opt-in in den Alarm-Einstellungen).
- [x] **Audit-Log-Aufbewahrung** (Pruning, doppelte Retention wie sonst).
