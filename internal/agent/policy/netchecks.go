package policy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

// strConfig liest einen String-Konfigwert (getrimmt).
func strConfig(c shared.CheckSpec, key string) string {
	s, _ := c.Config[key].(string)
	return strings.TrimSpace(s)
}

// timeoutDur liest den optionalen Timeout (Sekunden) oder nutzt den Default.
func timeoutDur(c shared.CheckSpec, def time.Duration) time.Duration {
	if v, ok := numConfig(c, "timeout"); ok && v > 0 {
		return time.Duration(v * float64(time.Second))
	}
	return def
}

// latencyStatus wendet optionale warn/crit-Schwellen (Millisekunden) auf einen
// Messwert an: über crit -> failing, über warn -> warning, sonst passing.
func latencyStatus(c shared.CheckSpec, ms float64) string {
	if crit, ok := numConfig(c, "crit"); ok && ms > crit {
		return "failing"
	}
	if warn, ok := numConfig(c, "warn"); ok && ms > warn {
		return "warning"
	}
	return "passing"
}

func unknownCheck(c shared.CheckSpec, msg string) shared.CheckResult {
	return shared.CheckResult{CheckID: c.ID, Status: "unknown", Output: msg}
}

// tcpCheck prüft, ob ein TCP-Port erreichbar ist (Verbindungsaufbau).
func tcpCheck(ctx context.Context, c shared.CheckSpec) shared.CheckResult {
	host := strConfig(c, "host")
	port, hasPort := numConfig(c, "port")
	if host == "" || !hasPort {
		return unknownCheck(c, "Host und Port erforderlich")
	}
	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
	d := net.Dialer{Timeout: timeoutDur(c, 5*time.Second)}
	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return shared.CheckResult{CheckID: c.ID, Status: "failing", Output: fmt.Sprintf("%s nicht erreichbar: %v", addr, err)}
	}
	_ = conn.Close()
	ms := float64(time.Since(start).Milliseconds())
	return shared.CheckResult{
		CheckID: c.ID, Status: latencyStatus(c, ms), Value: ms,
		Output: fmt.Sprintf("%s erreichbar in %.0f ms", addr, ms),
	}
}

// httpCheck prüft eine URL auf erwarteten Statuscode (Standard: 2xx/3xx) und
// optional die Antwortzeit.
func httpCheck(ctx context.Context, c shared.CheckSpec) shared.CheckResult {
	url := strConfig(c, "url")
	if url == "" {
		return unknownCheck(c, "URL erforderlich")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	client := &http.Client{Timeout: timeoutDur(c, 10*time.Second)}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return unknownCheck(c, "ungültige URL: "+err.Error())
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return shared.CheckResult{CheckID: c.ID, Status: "failing", Output: fmt.Sprintf("Anfrage an %s fehlgeschlagen: %v", url, err)}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	ms := float64(time.Since(start).Milliseconds())

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 400
	if want, ok := numConfig(c, "expected_status"); ok && want > 0 {
		statusOK = resp.StatusCode == int(want)
	}
	status := "failing"
	if statusOK {
		status = latencyStatus(c, ms) // erst bei richtigem Status zählt die Antwortzeit
	}
	return shared.CheckResult{
		CheckID: c.ID, Status: status, Value: float64(resp.StatusCode),
		Output: fmt.Sprintf("HTTP %d in %.0f ms (%s)", resp.StatusCode, ms, url),
	}
}

var pingTimeRe = regexp.MustCompile(`([0-9]+(?:[.,][0-9]+)?)\s*ms`)

// pingCheck prüft die Erreichbarkeit per System-Ping (ICMP) – ohne Raw-Socket-
// Rechte, plattformabhängige Argumente. Pass = antwortet; optional Latenz-Schwellen.
func pingCheck(ctx context.Context, c shared.CheckSpec) shared.CheckResult {
	host := strConfig(c, "host")
	if host == "" {
		return unknownCheck(c, "Host erforderlich")
	}
	to := timeoutDur(c, 5*time.Second)
	secs := int(to.Seconds())
	if secs < 1 {
		secs = 1
	}
	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"-n", "1", "-w", strconv.Itoa(int(to.Milliseconds())), host}
	case "darwin":
		args = []string{"-c", "1", "-t", strconv.Itoa(secs), host}
	default:
		args = []string{"-c", "1", "-W", strconv.Itoa(secs), host}
	}
	cctx, cancel := context.WithTimeout(ctx, to+2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cctx, "ping", args...).CombinedOutput()
	if err != nil {
		return shared.CheckResult{CheckID: c.ID, Status: "failing", Output: fmt.Sprintf("%s nicht erreichbar", host)}
	}
	ms, has := parsePingMs(string(out))
	status := "passing"
	if has {
		status = latencyStatus(c, ms)
	}
	val := 0.0
	suffix := ""
	if has {
		val = ms
		suffix = fmt.Sprintf(" (%.0f ms)", ms)
	}
	return shared.CheckResult{CheckID: c.ID, Status: status, Value: val, Output: fmt.Sprintf("%s erreichbar%s", host, suffix)}
}

// parsePingMs extrahiert die Antwortzeit (ms) aus einer Ping-Ausgabe (locale-tolerant).
func parsePingMs(out string) (float64, bool) {
	m := pingTimeRe.FindStringSubmatch(out)
	if len(m) < 2 {
		return 0, false
	}
	f, err := strconv.ParseFloat(strings.Replace(m[1], ",", ".", 1), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}
