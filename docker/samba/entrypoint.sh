#!/bin/bash
# Provisioniert beim ersten Start eine Samba-AD-Domäne und startet den DC
# anschließend im Vordergrund. Bei vorhandener Provisionierung wird nur gestartet.
set -e

REALM="${REALM:-EXAMPLE.LOCAL}"
DOMAIN="${DOMAIN:-EXAMPLE}"
ADMIN_PASS="${ADMIN_PASS:-Passw0rd!}"

if [ ! -f /var/lib/samba/private/sam.ldb ]; then
    echo "==> Provisioniere Samba-AD-Domäne ${REALM} (${DOMAIN}) ..."
    rm -f /etc/samba/smb.conf
    samba-tool domain provision \
        --use-rfc2307 \
        --domain="${DOMAIN}" \
        --realm="${REALM}" \
        --server-role=dc \
        --dns-backend=SAMBA_INTERNAL \
        --adminpass="${ADMIN_PASS}"

    cp /var/lib/samba/private/krb5.conf /etc/krb5.conf

    # Für bequemes Testen: Passwortrichtlinie entschärfen und Testbenutzer anlegen.
    samba-tool domain passwordsettings set --complexity=off --min-pwd-length=1 --history-length=0 --max-pwd-age=0 || true
    samba-tool user create tester "Test1234" --given-name=Test --surname=User || true
    samba-tool user create viewer "View1234" --given-name=View --surname=Only   || true
    samba-tool group add inventory-admins || true
    samba-tool group addmembers inventory-admins tester || true
    echo "==> Provisionierung abgeschlossen. Benutzer: tester/Test1234, viewer/View1234"
else
    echo "==> Bestehende Domäne gefunden, starte DC."
    [ -f /var/lib/samba/private/krb5.conf ] && cp /var/lib/samba/private/krb5.conf /etc/krb5.conf
fi

exec samba -i --debug-stdout
