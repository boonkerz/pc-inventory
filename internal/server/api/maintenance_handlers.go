package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/thomaspeterson/pc-inventory/internal/server/model"
)

func (s *Server) handleListMaintenance(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListMaintenanceWindows(r.Context())
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, items)
}

type maintenanceRequest struct {
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Note       string    `json:"note"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
}

func (s *Server) handleCreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var req maintenanceRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	switch req.TargetType {
	case "client", "site", "device":
	default:
		s.writeErr(w, http.StatusBadRequest, "target_type muss client, site oder device sein")
		return
	}
	if req.TargetID == "" {
		s.writeErr(w, http.StatusBadRequest, "target_id fehlt")
		return
	}
	if !req.EndsAt.After(req.StartsAt) {
		s.writeErr(w, http.StatusBadRequest, "Ende muss nach dem Beginn liegen")
		return
	}
	m := &model.MaintenanceWindow{
		TargetType: req.TargetType, TargetID: req.TargetID, Note: req.Note,
		StartsAt: req.StartsAt, EndsAt: req.EndsAt,
	}
	if err := s.store.CreateMaintenanceWindow(r.Context(), m); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, m)
}

func (s *Server) handleDeleteMaintenance(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteMaintenanceWindow(r.Context(), chi.URLParam(r, "id")); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
