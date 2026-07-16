package api

import (
	"testing"
	"time"
)

func TestFileCapHubResolveAndScope(t *testing.T) {
	h := &fileCapHub{caps: map[string]fileCap{}}
	h.add("tok-a", "dev-1")

	// Gültiges Token → Gerät.
	if dev, ok := h.resolve("tok-a"); !ok || dev != "dev-1" {
		t.Fatalf("resolve(tok-a) = %q,%v; want dev-1,true", dev, ok)
	}
	// Unbekanntes/leeres Token → abgelehnt.
	if _, ok := h.resolve("tok-x"); ok {
		t.Error("unbekanntes Token wurde akzeptiert")
	}
	if _, ok := h.resolve(""); ok {
		t.Error("leeres Token wurde akzeptiert")
	}
}

func TestFileCapHubExpiry(t *testing.T) {
	h := &fileCapHub{caps: map[string]fileCap{}}
	// Bereits abgelaufen.
	h.caps["old"] = fileCap{deviceID: "dev-1", expiry: time.Now().Add(-time.Second)}
	if _, ok := h.resolve("old"); ok {
		t.Error("abgelaufenes Token wurde akzeptiert")
	}
	// Gültig: resolve verlängert die Frist (sliding).
	h.caps["live"] = fileCap{deviceID: "dev-2", expiry: time.Now().Add(time.Second)}
	if _, ok := h.resolve("live"); !ok {
		t.Fatal("gültiges Token abgelehnt")
	}
	if got := h.caps["live"].expiry; time.Until(got) < 10*time.Minute {
		t.Errorf("resolve hat die Frist nicht verlängert: %v", got)
	}
}
