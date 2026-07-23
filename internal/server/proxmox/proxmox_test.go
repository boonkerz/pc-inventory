package proxmox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boonkerz/roster/internal/server/model"
)

func testHost(url string) *model.ProxmoxHost {
	return &model.ProxmoxHost{
		BaseURL: url, TokenID: "root@pam!roster", TokenSecret: "sekret", VerifyTLS: false,
	}
}

func TestListGuests(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api2/json/cluster/resources" || r.URL.Query().Get("type") != "vm" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[
			{"node":"pve","vmid":101,"type":"lxc","name":"web","status":"running"},
			{"node":"pve","vmid":102,"type":"qemu","name":"db","status":"stopped"},
			{"node":"pve","vmid":0,"type":"storage","name":"local","status":"available"}
		]}`))
	}))
	defer srv.Close()

	guests, err := New(testHost(srv.URL)).ListGuests(context.Background())
	if err != nil {
		t.Fatalf("ListGuests: %v", err)
	}
	if len(guests) != 2 { // storage-Eintrag wird gefiltert
		t.Fatalf("erwartete 2 Gäste, bekam %d: %+v", len(guests), guests)
	}
	if guests[0].Type != "lxc" || guests[0].VMID != 101 || guests[0].Name != "web" {
		t.Errorf("erster Gast falsch: %+v", guests[0])
	}
	if gotAuth != "PVEAPIToken=root@pam!roster=sekret" {
		t.Errorf("Auth-Header falsch: %q", gotAuth)
	}
}

func TestReboot(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":"UPID:..."}`))
	}))
	defer srv.Close()

	if err := New(testHost(srv.URL)).Reboot(context.Background(), "pve", "lxc", 101); err != nil {
		t.Fatalf("Reboot: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api2/json/nodes/pve/lxc/101/status/reboot" {
		t.Errorf("falscher Aufruf: %s %s", gotMethod, gotPath)
	}
	// ungültiger Typ -> Fehler ohne HTTP-Aufruf
	if err := New(testHost(srv.URL)).Reboot(context.Background(), "pve", "bogus", 1); err == nil {
		t.Error("ungültiger Gast-Typ sollte Fehler liefern")
	}
}
