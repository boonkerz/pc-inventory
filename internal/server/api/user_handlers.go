package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/thomaspeterson/pc-inventory/internal/server/auth"
	"github.com/thomaspeterson/pc-inventory/internal/server/model"
	"github.com/thomaspeterson/pc-inventory/internal/server/store"
)

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, users)
}

type createUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// handleCreateUser legt einen lokalen Benutzer an (nur Admin).
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	if req.Username == "" || req.Password == "" {
		s.writeErr(w, http.StatusBadRequest, "username und password erforderlich")
		return
	}
	role := model.Role(req.Role)
	if role != model.RoleAdmin && role != model.RoleTech && role != model.RoleViewer {
		role = model.RoleViewer
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		s.writeErr(w, http.StatusInternalServerError, "hashing fehlgeschlagen")
		return
	}
	u := &model.User{
		ID:           store.NewID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         role,
		AuthSource:   model.AuthLocal,
	}
	if err := s.store.CreateUser(r.Context(), u); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, u)
}

// handleAdminReset2FA deaktiviert die Zwei-Faktor-Authentifizierung eines Benutzers
// (Admin, für ausgesperrte Nutzer). Bei 2FA-Pflicht muss der Nutzer beim nächsten
// Login neu einrichten.
func (s *Server) handleAdminReset2FA(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := s.store.GetUserByID(r.Context(), id); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	if err := s.store.ClearUserTOTP(r.Context(), id); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
