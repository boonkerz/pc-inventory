package collect

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// PublicIP ermittelt die öffentliche IP-Adresse über einen externen Echo-Dienst.
// Best-effort: bei Fehler/offline wird "" zurückgegeben.
func PublicIP(ctx context.Context) string {
	client := &http.Client{Timeout: 6 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return ""
	}
	ip := strings.TrimSpace(string(b))
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
