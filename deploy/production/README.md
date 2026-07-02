# Produktions-Deployment (your-server:80 hinter eigenem Reverse-Proxy)

Der Server läuft als **HTTP auf :80** auf `your-server`. TLS für
`https://inventory.example.com` macht dein vorgelagerter Reverse-Proxy
(selbst eingerichtet). Wichtig: der Proxy muss **WebSockets** durchreichen
(Remote-Terminal, Wake-Long-Poll) und braucht **lange Timeouts** (≥ 60s, besser 1h).

## Server starten

Variante A – Repo liegt auf your-server (Docker baut dort):

```sh
docker compose -f deploy/production/compose.yml up -d --build
docker compose -f deploy/production/compose.yml logs inventory | grep -i passwort
```

Variante B – Image hier bauen und übertragen (kein Build auf dem Zielhost):

```sh
# auf DIESER Maschine:
docker build -f docker/server/Dockerfile -t pcinv-server:0.1.0 --build-arg VERSION=0.1.0 .
docker save pcinv-server:0.1.0 | gzip > pcinv-server.tgz
scp pcinv-server.tgz user@your-server:~/

# auf your-server:
gunzip -c pcinv-server.tgz | docker load
docker volume create inventory-data
docker run -d --name pcinv --restart unless-stopped -p 80:8443 \
  -e PCINV_ADDR=":8443" -e PCINV_DB="sqlite:///var/lib/pc-inventory/inventory.db" \
  -e PCINV_BEHIND_PROXY=true -e PCINV_SECURE_COOKIE=true -e PCINV_REQUIRE_2FA=true \
  -e PCINV_SEED_ADMIN_USER=admin \
  -v inventory-data:/var/lib/pc-inventory pcinv-server:0.1.0
docker logs pcinv | grep -i passwort   # einmaliges Admin-Passwort
```

## Wichtige Env-Variablen

- `PCINV_BEHIND_PROXY=true` – behält Secure-Cookies, obwohl der Server selbst HTTP
  spricht (TLS macht der Proxy). **Ohne das kommst du nicht rein.**
- `PCINV_SECURE_COOKIE=true`, `PCINV_REQUIRE_2FA=true` (Default).
- **Kein** `PCINV_SEED_ENROLL_TOKEN` in Produktion – Enrollment-Tokens in der UI erzeugen.
- Admin-Passwort: `PCINV_SEED_ADMIN_PASSWORD` leer lassen → wird beim Erststart erzeugt
  und einmalig geloggt.

## Erstinbetriebnahme

1. `https://inventory.example.com` öffnen, `admin` + geloggtes Passwort.
2. **2FA einrichten** (Pflicht), Backup-Codes sichern, ggf. Passwort ändern.
3. Einstellungen → **Enrollment-Token** erzeugen.
4. Agents zeigen auf `https://inventory.example.com` mit normaler
   Zertifikatsprüfung (kein `insecure_skip_verify`/Pinning – der Proxy hat ein echtes Cert).

`:80` auf your-server nicht direkt ins Internet exponieren – nur der Proxy spricht damit.
