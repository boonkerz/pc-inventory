package policy

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boonkerz/roster/internal/shared"
)

func TestTCPCheck(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	var p float64
	{ // port als float64 wie aus JSON
		var n int
		for _, r := range port { n = n*10 + int(r-'0') }
		p = float64(n)
	}
	c := shared.CheckSpec{ID: "1", Type: "tcp", Config: map[string]any{"host": "127.0.0.1", "port": p}}
	if r := tcpCheck(context.Background(), c); r.Status != "passing" {
		t.Fatalf("offener Port sollte passing sein: %+v", r)
	}
	// geschlossener Port
	c.Config["port"] = float64(1) // unwahrscheinlich offen
	if r := tcpCheck(context.Background(), c); r.Status != "failing" {
		t.Fatalf("geschlossener Port sollte failing sein: %+v", r)
	}
}

func TestHTTPCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := shared.CheckSpec{ID: "1", Type: "http", Config: map[string]any{"url": srv.URL}}
	if r := httpCheck(context.Background(), c); r.Status != "passing" || r.Value != 200 {
		t.Fatalf("200 sollte passing sein: %+v", r)
	}
	// erwartet 404 -> failing
	c.Config["expected_status"] = float64(404)
	if r := httpCheck(context.Background(), c); r.Status != "failing" {
		t.Fatalf("Status-Mismatch sollte failing sein: %+v", r)
	}
	// unerreichbar
	c2 := shared.CheckSpec{ID: "2", Type: "http", Config: map[string]any{"url": "http://127.0.0.1:1"}}
	if r := httpCheck(context.Background(), c2); r.Status != "failing" {
		t.Fatalf("unerreichbar sollte failing sein: %+v", r)
	}
}

func TestHTTPCheckContains(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("<html><body>Willkommen bei Roster</body></html>"))
	}))
	defer srv.Close()

	// Text vorhanden -> passing
	c := shared.CheckSpec{ID: "1", Type: "http", Config: map[string]any{"url": srv.URL, "contains": "Willkommen"}}
	if r := httpCheck(context.Background(), c); r.Status != "passing" {
		t.Fatalf("vorhandener Text sollte passing sein: %+v", r)
	}
	// Text fehlt -> failing, obwohl HTTP 200
	c.Config["contains"] = "Fehlermeldung"
	if r := httpCheck(context.Background(), c); r.Status != "failing" {
		t.Fatalf("fehlender Text sollte failing sein: %+v", r)
	}
}

func TestParsePingMs(t *testing.T) {
	for _, s := range []string{"64 bytes from x: icmp_seq=0 ttl=64 time=0.045 ms", "Antwort von 127.0.0.1: Zeit=1ms"} {
		if _, ok := parsePingMs(s); !ok {
			t.Fatalf("Latenz sollte erkannt werden in %q", s)
		}
	}
	if _, ok := parsePingMs("keine zahl hier"); ok {
		t.Fatal("sollte nichts finden")
	}
}
