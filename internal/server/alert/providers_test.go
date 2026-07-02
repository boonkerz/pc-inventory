package alert

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// captureHTTP ersetzt den Paket-HTTP-Client durch einen, der die Anfrage abfängt.
func captureHTTP(t *testing.T) (get func() (*http.Request, []byte)) {
	t.Helper()
	var lastReq *http.Request
	var lastBody []byte
	orig := httpClient
	httpClient = &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		lastReq = r
		if r.Body != nil {
			lastBody, _ = io.ReadAll(r.Body)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok")), Header: http.Header{}}, nil
	})}
	t.Cleanup(func() { httpClient = orig })
	return func() (*http.Request, []byte) { return lastReq, lastBody }
}

func TestWebhookProvider(t *testing.T) {
	get := captureHTTP(t)
	err := webhookProvider{}.Send(context.Background(),
		map[string]string{"url": "https://example.test/hook"},
		Notification{Subject: "S", Body: "B"})
	if err != nil {
		t.Fatal(err)
	}
	req, body := get()
	if req.Method != http.MethodPost || req.URL.String() != "https://example.test/hook" {
		t.Fatalf("falsche request: %s %s", req.Method, req.URL)
	}
	var m map[string]string
	if json.Unmarshal(body, &m); m["subject"] != "S" || m["body"] != "B" {
		t.Fatalf("falscher body: %s", body)
	}
}

func TestPushoverProvider(t *testing.T) {
	get := captureHTTP(t)
	err := pushoverProvider{}.Send(context.Background(),
		map[string]string{"token": "tok", "user": "usr"},
		Notification{Subject: "Titel", Body: "Text"})
	if err != nil {
		t.Fatal(err)
	}
	req, body := get()
	if req.URL.Host != "api.pushover.net" {
		t.Fatalf("falscher host: %s", req.URL)
	}
	form, _ := url.ParseQuery(string(body))
	if form.Get("token") != "tok" || form.Get("user") != "usr" || form.Get("title") != "Titel" || form.Get("message") != "Text" {
		t.Fatalf("falsches form: %v", form)
	}
}

func TestTelegramProvider(t *testing.T) {
	get := captureHTTP(t)
	err := telegramProvider{}.Send(context.Background(),
		map[string]string{"bot_token": "123:abc", "chat_id": "42"},
		Notification{Subject: "S", Body: "B"})
	if err != nil {
		t.Fatal(err)
	}
	req, body := get()
	if !strings.Contains(req.URL.Path, "/bot123:abc/sendMessage") {
		t.Fatalf("falscher pfad: %s", req.URL.Path)
	}
	form, _ := url.ParseQuery(string(body))
	if form.Get("chat_id") != "42" || !strings.Contains(form.Get("text"), "S") {
		t.Fatalf("falsches form: %v", form)
	}
}

func TestNtfyProvider(t *testing.T) {
	get := captureHTTP(t)
	err := ntfyProvider{}.Send(context.Background(),
		map[string]string{"server": "https://ntfy.example", "topic": "alarme", "token": "secret"},
		Notification{Subject: "Titel", Body: "Inhalt"})
	if err != nil {
		t.Fatal(err)
	}
	req, body := get()
	if req.URL.String() != "https://ntfy.example/alarme" {
		t.Fatalf("falsche url: %s", req.URL)
	}
	if req.Header.Get("Title") != "Titel" || string(body) != "Inhalt" {
		t.Fatalf("falscher title/body: %q / %q", req.Header.Get("Title"), body)
	}
	if req.Header.Get("Authorization") != "Bearer secret" {
		t.Fatalf("token fehlt: %q", req.Header.Get("Authorization"))
	}
}

func TestRequiredFieldErrors(t *testing.T) {
	cases := []struct {
		p   Provider
		cfg map[string]string
	}{
		{webhookProvider{}, map[string]string{}},
		{pushoverProvider{}, map[string]string{"token": "x"}},
		{telegramProvider{}, map[string]string{"chat_id": "1"}},
		{ntfyProvider{}, map[string]string{}},
		{emailProvider{}, map[string]string{}},
	}
	for _, c := range cases {
		if err := c.p.Send(context.Background(), c.cfg, Notification{}); err == nil {
			t.Errorf("%s: erwartete Fehler bei unvollständiger Config", c.p.Type())
		}
	}
}

func TestCatalogAndSecretKeys(t *testing.T) {
	cat := Catalog()
	if len(cat) != 5 {
		t.Fatalf("erwartete 5 provider, %d", len(cat))
	}
	if !SecretKeys("email")["pass"] {
		t.Error("email.pass sollte geheim sein")
	}
	if !SecretKeys("pushover")["token"] {
		t.Error("pushover.token sollte geheim sein")
	}
	if SecretKeys("telegram")["chat_id"] {
		t.Error("telegram.chat_id sollte NICHT geheim sein")
	}
}
