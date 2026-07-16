package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoteJoin(t *testing.T) {
	cases := []struct{ dir, name, want string }{
		{"/home/thomas", "a.txt", "/home/thomas/a.txt"},
		{"/home/thomas/", "a.txt", "/home/thomas/a.txt"},
		{`C:\Users\t`, "a.txt", `C:\Users\t\a.txt`},
		{`C:\`, "a.txt", `C:\a.txt`},
		{"", "a.txt", "a.txt"},
	}
	for _, c := range cases {
		if got := remoteJoin(c.dir, c.name); got != c.want {
			t.Errorf("remoteJoin(%q,%q) = %q, want %q", c.dir, c.name, got, c.want)
		}
	}
}

func TestHTTPBase(t *testing.T) {
	cases := map[string]string{
		"https://x.de":     "https://x.de",
		"http://x.de:8080": "http://x.de:8080",
		"wss://x.de":       "https://x.de",
		"ws://x.de":        "http://x.de",
		"x.de":             "https://x.de",
	}
	for in, want := range cases {
		if got := httpBase(in); got != want {
			t.Errorf("httpBase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{0: "0B", 512: "512B", 2048: "2K", 3 << 20: "3.0M", 5 << 30: "5.0G"}
	for in, want := range cases {
		if got := humanSize(in); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", in, got, want)
		}
	}
}

// TestFMClientBrowse prüft den kompletten Client-Protokollfluss (queue → poll →
// FileListing) gegen einen Stub, der die viewer-files-API nachbildet.
func TestFMClientBrowse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/viewer-files/browse"):
			_ = json.NewEncoder(w).Encode(map[string]string{"command_id": "c1"})
		case strings.Contains(r.URL.Path, "/viewer-files/command/"):
			listing := `{"path":"/home","parent":"/","entries":[{"name":"a.txt","path":"/home/a.txt","dir":false,"size":10},{"name":"sub","path":"/home/sub","dir":true}]}`
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "done", "exit_code": 0, "output": listing})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cl := &fmClient{http: srv.Client(), base: srv.URL, device: "dev1", token: "tok"}
	l := cl.browse("/home")
	if l.err != "" {
		t.Fatalf("browse err: %s", l.err)
	}
	if l.path != "/home" || len(l.entries) != 2 {
		t.Fatalf("unerwartetes Listing: %+v", l)
	}
	if !l.entries[1].dir || l.entries[0].size != 10 {
		t.Errorf("Einträge falsch geparst: %+v", l.entries)
	}
}

func TestFMClientBrowseAuthFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	cl := &fmClient{http: srv.Client(), base: srv.URL, device: "dev1", token: "bad"}
	if l := cl.browse("/x"); l.err == "" {
		t.Error("erwartete einen Fehler bei 401")
	}
}
