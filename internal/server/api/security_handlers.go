package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// queueDeviceCmd ist ein kleiner Helfer für parameterlose Geräte-Befehle.
func (s *Server) queueDeviceCmd(w http.ResponseWriter, r *http.Request, typ, label string, payload map[string]any) {
	id, err := s.queueCommand(r.Context(), chi.URLParam(r, "id"), typ, label, payload)
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, map[string]string{"command_id": id})
}

func (s *Server) handleAVStatus(w http.ResponseWriter, r *http.Request) {
	s.queueDeviceCmd(w, r, "av_status", "Virenschutz-Status", nil)
}

func (s *Server) handleBitLockerStatus(w http.ResponseWriter, r *http.Request) {
	s.queueDeviceCmd(w, r, "bitlocker_status", "BitLocker-Status", nil)
}

func (s *Server) handleSmartStatus(w http.ResponseWriter, r *http.Request) {
	s.queueDeviceCmd(w, r, "smart_status", "Datenträgergesundheit (SMART)", nil)
}

type eventLogRequest struct {
	Log   string `json:"log"`
	Count int    `json:"count"`
}

func (s *Server) handleEventLog(w http.ResponseWriter, r *http.Request) {
	var req eventLogRequest
	if !s.decodeJSON(w, r, &req) {
		return
	}
	s.queueDeviceCmd(w, r, "event_log", "Ereignisprotokoll", map[string]any{"log": req.Log, "count": req.Count})
}

// handleBitLockerResult liefert das (gepollte) BitLocker-Ergebnis eines Befehls und
// escrowt dabei enthaltene Wiederherstellungsschlüssel serverseitig.
func (s *Server) handleBitLockerResult(w http.ResponseWriter, r *http.Request) {
	cmd, err := s.store.CommandByID(r.Context(), chi.URLParam(r, "cmd"))
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	if cmd.Status != "done" || cmd.Output == "" {
		s.writeJSON(w, http.StatusOK, map[string]string{"status": cmd.Status})
		return
	}
	var res struct {
		Volumes []struct {
			MountPoint  string `json:"mount_point"`
			Protection  string `json:"protection"`
			Percent     int    `json:"percent"`
			RecoveryKey string `json:"recovery_key"`
			RecoveryID  string `json:"recovery_id"`
		} `json:"volumes"`
		Info string `json:"info,omitempty"`
	}
	if err := json.Unmarshal([]byte(cmd.Output), &res); err != nil {
		s.writeJSON(w, http.StatusOK, json.RawMessage(cmd.Output))
		return
	}
	now := time.Now()
	for _, v := range res.Volumes {
		if v.RecoveryKey != "" {
			_ = s.store.SaveBitLockerKey(r.Context(), cmd.DeviceID, v.MountPoint, v.RecoveryID, v.RecoveryKey, now)
		}
	}
	s.writeJSON(w, http.StatusOK, res)
}
