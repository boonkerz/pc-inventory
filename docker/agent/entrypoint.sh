#!/bin/sh
# Schreibt die Agent-Konfiguration aus Umgebungsvariablen und startet den Agent
# im Vordergrund (kein Dienst-Manager im Container).
set -e

: "${ROSTER_SERVER_URL:?ROSTER_SERVER_URL muss gesetzt sein}"
: "${ROSTER_ENROLLMENT_TOKEN:?ROSTER_ENROLLMENT_TOKEN muss gesetzt sein}"
INTERVAL="${ROSTER_INTERVAL:-30s}"
INSECURE="${ROSTER_INSECURE:-true}"
CACERT="${ROSTER_CACERT:-}"

# Auf das (vom Server erzeugte) CA-Zertifikat warten, falls eines gepinnt werden soll.
if [ -n "$CACERT" ]; then
  echo "warte auf CA-Zertifikat ${CACERT} ..."
  i=0
  while [ "$i" -lt 60 ] && [ ! -f "$CACERT" ]; do i=$((i+1)); sleep 1; done
fi

mkdir -p /etc/roster /var/lib/roster
cat > /etc/roster/agent.yaml <<EOF
server_url: "${ROSTER_SERVER_URL}"
enrollment_token: "${ROSTER_ENROLLMENT_TOKEN}"
insecure_skip_verify: ${INSECURE}
ca_cert_path: "${CACERT}"
interval: "${INTERVAL}"
state_path: "/var/lib/roster/agent-state.json"
EOF

# Auf den Server warten (Root-URL liefert 200), damit das erste Enrollment klappt.
echo "warte auf ${ROSTER_SERVER_URL} ..."
i=0
while [ "$i" -lt 60 ]; do
  if wget -q -O /dev/null --no-check-certificate "${ROSTER_SERVER_URL}/" 2>/dev/null; then
    echo "server erreichbar"; break
  fi
  i=$((i+1)); sleep 2
done

exec agent -config /etc/roster/agent.yaml run
