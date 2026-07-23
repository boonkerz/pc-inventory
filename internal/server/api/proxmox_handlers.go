package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/boonkerz/roster/internal/server/model"
	"github.com/boonkerz/roster/internal/server/proxmox"
	"github.com/boonkerz/roster/internal/server/store"
)

// handleListProxmoxHosts liefert alle Proxmox-Hosts (ohne Token-Secret).
func (s *Server) handleListProxmoxHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := s.store.ListProxmoxHosts(r.Context())
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	if hosts == nil {
		hosts = []model.ProxmoxHost{}
	}
	s.writeJSON(w, http.StatusOK, hosts)
}

type proxmoxHostRequest struct {
	Name        string `json:"name"`
	BaseURL     string `json:"base_url"`
	TokenID     string `json:"token_id"`
	TokenSecret string `json:"token_secret"`
	VerifyTLS   bool   `json:"verify_tls"`
}

// handleCreateProxmoxHost legt einen Proxmox-Host an.
func (s *Server) handleCreateProxmoxHost(w http.ResponseWriter, r *http.Request) {
	var req proxmoxHostRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.TokenID = strings.TrimSpace(req.TokenID)
	req.TokenSecret = strings.TrimSpace(req.TokenSecret)
	if req.BaseURL == "" || req.TokenID == "" || req.TokenSecret == "" {
		s.writeErr(w, http.StatusBadRequest, "base_url, token_id und token_secret erforderlich")
		return
	}
	if !strings.Contains(req.TokenID, "!") {
		s.writeErr(w, http.StatusBadRequest, "token_id muss das Format user@realm!tokenname haben")
		return
	}
	h := &model.ProxmoxHost{
		ID: store.NewID(), Name: req.Name, BaseURL: req.BaseURL,
		TokenID: req.TokenID, TokenSecret: req.TokenSecret, VerifyTLS: req.VerifyTLS,
	}
	if err := s.store.CreateProxmoxHost(r.Context(), h); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	h.TokenSecret = "" // nie zurückgeben
	s.writeJSON(w, http.StatusCreated, h)
}

// handleDeleteProxmoxHost entfernt einen Proxmox-Host.
func (s *Server) handleDeleteProxmoxHost(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteProxmoxHost(r.Context(), chi.URLParam(r, "id")); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "gelöscht"})
}

// handleListProxmoxGuests fragt die Gäste (LXC/QEMU) eines Hosts live von Proxmox ab.
func (s *Server) handleListProxmoxGuests(w http.ResponseWriter, r *http.Request) {
	host, err := s.store.GetProxmoxHost(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	guests, err := proxmox.New(host).ListGuests(r.Context())
	if err != nil {
		s.writeErr(w, http.StatusBadGateway, "Proxmox nicht erreichbar: "+err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, guests)
}

type proxmoxRebootRequest struct {
	Node string `json:"node"`
	Type string `json:"type"`
	VMID int    `json:"vmid"`
}

// handleRebootProxmoxGuest rebootet einen Gast manuell (Knopf in der UI).
func (s *Server) handleRebootProxmoxGuest(w http.ResponseWriter, r *http.Request) {
	host, err := s.store.GetProxmoxHost(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	var req proxmoxRebootRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	if err := proxmox.New(host).Reboot(r.Context(), req.Node, req.Type, req.VMID); err != nil {
		s.writeErr(w, http.StatusBadGateway, "Reboot fehlgeschlagen: "+err.Error())
		return
	}
	uname := ""
	if u := userFrom(r.Context()); u != nil {
		uname = u.Username
	}
	_ = s.store.InsertAudit(r.Context(), model.AuditEntry{
		TS: time.Now().UTC(), Username: uname,
		Action: "Proxmox-Gast rebootet (" + host.Name + " " + req.Type + "/" + strconv.Itoa(req.VMID) + ")",
		Method: "POST", Path: r.URL.Path, Status: 200,
	})
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "reboot ausgelöst"})
}
