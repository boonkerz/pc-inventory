#!/bin/sh
# Schreibt die Agent-Konfiguration aus Umgebungsvariablen und startet den Agent
# im Vordergrund (kein Dienst-Manager im Container).
set -e

: "${PCINV_SERVER_URL:?PCINV_SERVER_URL muss gesetzt sein}"
: "${PCINV_ENROLLMENT_TOKEN:?PCINV_ENROLLMENT_TOKEN muss gesetzt sein}"
INTERVAL="${PCINV_INTERVAL:-30s}"
INSECURE="${PCINV_INSECURE:-true}"
CACERT="${PCINV_CACERT:-}"

# Auf das (vom Server erzeugte) CA-Zertifikat warten, falls eines gepinnt werden soll.
if [ -n "$CACERT" ]; then
  echo "warte auf CA-Zertifikat ${CACERT} ..."
  i=0
  while [ "$i" -lt 60 ] && [ ! -f "$CACERT" ]; do i=$((i+1)); sleep 1; done
fi

mkdir -p /etc/pc-inventory /var/lib/pc-inventory
cat > /etc/pc-inventory/agent.yaml <<EOF
server_url: "${PCINV_SERVER_URL}"
enrollment_token: "${PCINV_ENROLLMENT_TOKEN}"
insecure_skip_verify: ${INSECURE}
ca_cert_path: "${CACERT}"
interval: "${INTERVAL}"
state_path: "/var/lib/pc-inventory/agent-state.json"
EOF

# Auf den Server warten (Root-URL liefert 200), damit das erste Enrollment klappt.
echo "warte auf ${PCINV_SERVER_URL} ..."
i=0
while [ "$i" -lt 60 ]; do
  if wget -q -O /dev/null --no-check-certificate "${PCINV_SERVER_URL}/" 2>/dev/null; then
    echo "server erreichbar"; break
  fi
  i=$((i+1)); sleep 2
done

exec agent -config /etc/pc-inventory/agent.yaml run
