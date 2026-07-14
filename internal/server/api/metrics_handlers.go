package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/boonkerz/roster/internal/server/store"
)

// handleMetricsHistory liefert die Auslastungs-Historie eines Geräts, je nach
// Zeitraum passend gemittelt (24h/7d/30d).
func (s *Server) handleMetricsHistory(w http.ResponseWriter, r *http.Request) {
	var window, bucket time.Duration
	switch r.URL.Query().Get("range") {
	case "7d":
		window, bucket = 7*24*time.Hour, 30*time.Minute
	case "30d":
		window, bucket = 30*24*time.Hour, 2*time.Hour
	default: // 24h
		window, bucket = 24*time.Hour, 5*time.Minute
	}
	since := time.Now().Add(-window).UnixMilli()
	points, err := s.store.MetricsHistory(r.Context(), chi.URLParam(r, "id"), since, bucket.Milliseconds())
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	if points == nil {
		points = []store.MetricPoint{}
	}
	s.writeJSON(w, http.StatusOK, points)
}
