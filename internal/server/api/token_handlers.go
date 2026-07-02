package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/thomaspeterson/pc-inventory/internal/server/auth"
	"github.com/thomaspeterson/pc-inventory/internal/server/model"
	"github.com/thomaspeterson/pc-inventory/internal/server/store"
)

func (s *Server) handleListTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := s.store.ListEnrollmentTokens(r.Context())
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, tokens)
}

type createTokenRequest struct {
	Label      string  `json:"label"`
	MaxUses    int     `json:"max_uses"`
	ExpiresInH int     `json:"expires_in_hours"` // 0 = nie
	SiteID     *string `json:"site_id"`          // optional: Token an Standort binden
}

// handleCreateToken erzeugt ein Enrollment-Token und gibt es EINMALIG im Klartext zurück.
func (s *Server) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	user := userFrom(r.Context())
	plain := auth.GenerateToken()
	tok := &model.EnrollmentToken{
		ID:        store.NewID(),
		Label:     req.Label,
		MaxUses:   req.MaxUses,
		CreatedBy: user.Username,
		SiteID:    req.SiteID,
	}
	if req.ExpiresInH > 0 {
		exp := time.Now().Add(time.Duration(req.ExpiresInH) * time.Hour).UTC()
		tok.ExpiresAt = &exp
	}
	if err := s.store.CreateEnrollmentToken(r.Context(), tok, auth.HashToken(plain)); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	tok.Token = plain // nur in dieser Antwort
	s.writeJSON(w, http.StatusCreated, tok)
}

func (s *Server) handleDeleteToken(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteEnrollmentToken(r.Context(), chi.URLParam(r, "id")); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "gelöscht"})
}
