package alert

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// --- E-Mail (SMTP) ---

type emailProvider struct{}

func (emailProvider) Type() string  { return "email" }
func (emailProvider) Label() string { return "E-Mail (SMTP)" }
func (emailProvider) Fields() []Field {
	return []Field{
		{Key: "host", Label: "SMTP-Host", Type: "text", Required: true},
		{Key: "port", Label: "Port", Type: "number", Help: "587 = STARTTLS, 465 = implizites TLS"},
		{Key: "user", Label: "Benutzer", Type: "text"},
		{Key: "pass", Label: "Passwort", Type: "password"},
		{Key: "from", Label: "Absender (From)", Type: "text"},
		{Key: "recipient", Label: "Empfänger (kommagetrennt)", Type: "text", Required: true},
	}
}

func (emailProvider) Send(_ context.Context, cfg map[string]string, n Notification) error {
	host := cfg["host"]
	if host == "" || cfg["recipient"] == "" {
		return fmt.Errorf("host und empfänger erforderlich")
	}
	port, _ := strconv.Atoi(cfg["port"])
	if port == 0 {
		port = 587
	}
	from := cfg["from"]
	if from == "" {
		from = cfg["user"]
	}
	to := splitList(cfg["recipient"])
	msg := buildMessage(from, to, n.Subject, n.Body)
	addr := fmt.Sprintf("%s:%d", host, port)

	var auth smtp.Auth
	if cfg["user"] != "" {
		auth = smtp.PlainAuth("", cfg["user"], cfg["pass"], host)
	}
	if port == 465 {
		return sendImplicitTLS(addr, host, auth, from, to, msg)
	}
	return smtp.SendMail(addr, auth, from, to, msg)
}

func sendImplicitTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Quit()
	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(strings.TrimSpace(rcpt)); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}

func buildMessage(from string, to []string, subject, body string) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(body)
	return b.Bytes()
}

// --- Generischer Webhook ---

type webhookProvider struct{}

func (webhookProvider) Type() string  { return "webhook" }
func (webhookProvider) Label() string { return "Webhook (JSON)" }
func (webhookProvider) Fields() []Field {
	return []Field{
		{Key: "url", Label: "URL", Type: "text", Required: true, Help: "POST mit JSON {subject, body}. Auch für Slack/Discord/Teams-Webhooks."},
	}
}

func (webhookProvider) Send(ctx context.Context, cfg map[string]string, n Notification) error {
	if cfg["url"] == "" {
		return fmt.Errorf("url erforderlich")
	}
	payload, _ := json.Marshal(map[string]string{"subject": n.Subject, "body": n.Body})
	return postJSON(ctx, cfg["url"], payload, nil)
}

// --- Pushover ---

type pushoverProvider struct{}

func (pushoverProvider) Type() string  { return "pushover" }
func (pushoverProvider) Label() string { return "Pushover" }
func (pushoverProvider) Fields() []Field {
	return []Field{
		{Key: "token", Label: "API-Token/Key", Type: "password", Required: true},
		{Key: "user", Label: "User-/Group-Key", Type: "text", Required: true},
		{Key: "priority", Label: "Priorität (-2..2)", Type: "number", Help: "optional, Standard 0"},
	}
}

func (pushoverProvider) Send(ctx context.Context, cfg map[string]string, n Notification) error {
	if cfg["token"] == "" || cfg["user"] == "" {
		return fmt.Errorf("token und user erforderlich")
	}
	form := url.Values{
		"token":   {cfg["token"]},
		"user":    {cfg["user"]},
		"title":   {n.Subject},
		"message": {n.Body},
	}
	if p := cfg["priority"]; p != "" {
		form.Set("priority", p)
	}
	return postForm(ctx, "https://api.pushover.net/1/messages.json", form)
}

// --- Telegram ---

type telegramProvider struct{}

func (telegramProvider) Type() string  { return "telegram" }
func (telegramProvider) Label() string { return "Telegram" }
func (telegramProvider) Fields() []Field {
	return []Field{
		{Key: "bot_token", Label: "Bot-Token", Type: "password", Required: true},
		{Key: "chat_id", Label: "Chat-ID", Type: "text", Required: true},
	}
}

func (telegramProvider) Send(ctx context.Context, cfg map[string]string, n Notification) error {
	if cfg["bot_token"] == "" || cfg["chat_id"] == "" {
		return fmt.Errorf("bot_token und chat_id erforderlich")
	}
	text := n.Subject
	if n.Body != "" {
		text += "\n\n" + n.Body
	}
	form := url.Values{"chat_id": {cfg["chat_id"]}, "text": {text}}
	endpoint := "https://api.telegram.org/bot" + cfg["bot_token"] + "/sendMessage"
	return postForm(ctx, endpoint, form)
}

// --- ntfy ---

type ntfyProvider struct{}

func (ntfyProvider) Type() string  { return "ntfy" }
func (ntfyProvider) Label() string { return "ntfy" }
func (ntfyProvider) Fields() []Field {
	return []Field{
		{Key: "server", Label: "Server", Type: "text", Help: "Standard: https://ntfy.sh"},
		{Key: "topic", Label: "Topic", Type: "text", Required: true},
		{Key: "token", Label: "Access-Token", Type: "password", Help: "optional (für geschützte Topics)"},
	}
}

func (ntfyProvider) Send(ctx context.Context, cfg map[string]string, n Notification) error {
	if cfg["topic"] == "" {
		return fmt.Errorf("topic erforderlich")
	}
	server := cfg["server"]
	if server == "" {
		server = "https://ntfy.sh"
	}
	endpoint := strings.TrimRight(server, "/") + "/" + cfg["topic"]
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(n.Body))
	if err != nil {
		return err
	}
	if n.Subject != "" {
		req.Header.Set("Title", n.Subject)
	}
	if cfg["token"] != "" {
		req.Header.Set("Authorization", "Bearer "+cfg["token"])
	}
	return doExpectOK(req)
}

// --- HTTP-Helfer ---

func postJSON(ctx context.Context, endpoint string, body []byte, header http.Header) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	return doExpectOK(req)
}

func postForm(ctx context.Context, endpoint string, form url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return doExpectOK(req)
}

func doExpectOK(req *http.Request) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP-Status %d", resp.StatusCode)
	}
	return nil
}

func splitList(s string) []string {
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
