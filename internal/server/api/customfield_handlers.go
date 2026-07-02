package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/thomaspeterson/pc-inventory/internal/server/model"
	"github.com/thomaspeterson/pc-inventory/internal/server/store"
)

var validFieldTypes = map[string]bool{
	"text": true, "number": true, "checkbox": true, "select": true,
	"multiselect": true, "datetime": true, "list": true,
}
var validFieldModels = map[string]bool{"client": true, "site": true, "device": true}

// --- Definitionen ---

func (s *Server) handleListCustomFields(w http.ResponseWriter, r *http.Request) {
	fields, err := s.store.CustomFields(r.Context(), r.URL.Query().Get("model"))
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, fields)
}

type customFieldRequest struct {
	Model    string   `json:"model"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Options  []string `json:"options"`
	Default  string   `json:"default_value"`
	Required bool     `json:"required"`
}

func (s *Server) handleCreateCustomField(w http.ResponseWriter, r *http.Request) {
	var req customFieldRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" || !validFieldModels[req.Model] || !validFieldTypes[req.Type] {
		s.writeErr(w, http.StatusBadRequest, "name, gültiges model (client|site|device) und type erforderlich")
		return
	}
	f := &model.CustomField{
		ID: store.NewID(), Model: req.Model, Name: req.Name, Type: req.Type,
		Options: req.Options, Default: req.Default, Required: req.Required,
	}
	if err := s.store.CreateCustomField(r.Context(), f); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, f)
}

func (s *Server) handleUpdateCustomField(w http.ResponseWriter, r *http.Request) {
	var req customFieldRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	if req.Name == "" || !validFieldTypes[req.Type] {
		s.writeErr(w, http.StatusBadRequest, "name und gültiger type erforderlich")
		return
	}
	f := &model.CustomField{
		ID: chi.URLParam(r, "id"), Name: req.Name, Type: req.Type,
		Options: req.Options, Default: req.Default, Required: req.Required,
	}
	if err := s.store.UpdateCustomField(r.Context(), f); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "gespeichert"})
}

func (s *Server) handleDeleteCustomField(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteCustomField(r.Context(), chi.URLParam(r, "id")); err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "gelöscht"})
}

// --- Werte ---

func (s *Server) handleGetCustomFieldValues(w http.ResponseWriter, r *http.Request) {
	mdl := r.URL.Query().Get("model")
	entityID := r.URL.Query().Get("entity_id")
	if !validFieldModels[mdl] || entityID == "" {
		s.writeErr(w, http.StatusBadRequest, "model und entity_id erforderlich")
		return
	}
	vals, err := s.store.CustomFieldValues(r.Context(), mdl, entityID)
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, vals)
}

type setValuesRequest struct {
	Model    string         `json:"model"`
	EntityID string         `json:"entity_id"`
	Values   map[string]any `json:"values"` // field_id -> Wert (String oder Array)
}

func (s *Server) handleSetCustomFieldValues(w http.ResponseWriter, r *http.Request) {
	var req setValuesRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	if !validFieldModels[req.Model] || req.EntityID == "" {
		s.writeErr(w, http.StatusBadRequest, "model und entity_id erforderlich")
		return
	}
	for fieldID, raw := range req.Values {
		if err := s.store.SetCustomFieldValue(r.Context(), fieldID, req.EntityID, fieldValueToString(raw)); err != nil {
			s.mapStoreErr(w, err)
			return
		}
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "gespeichert"})
}

// fieldValueToString speichert Strings direkt, alles andere (Arrays/Zahlen/Bool) als JSON.
func fieldValueToString(v any) string {
	if str, ok := v.(string); ok {
		return str
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
