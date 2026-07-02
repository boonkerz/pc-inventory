package api

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

const maxFileTransfer = 32 << 20 // 32 MiB, passend zum Agent-Limit

// fileBlob ist eine zwischengelagerte Datei-Übertragung (Download vom Agent oder
// Upload vom Browser), transient im Speicher gehalten bis zur Abholung/TTL.
type fileBlob struct {
	name string
	data []byte
	ts   time.Time
}

// fileHub hält laufende Datei-Übertragungen (keyed by Transfer-/Command-ID).
type fileHub struct {
	mu    sync.Mutex
	blobs map[string]fileBlob
}

func newFileHub() *fileHub {
	h := &fileHub{blobs: map[string]fileBlob{}}
	go h.gc()
	return h
}

func (h *fileHub) put(id, name string, data []byte) {
	h.mu.Lock()
	h.blobs[id] = fileBlob{name: name, data: data, ts: time.Now()}
	h.mu.Unlock()
}

func (h *fileHub) take(id string) (fileBlob, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	b, ok := h.blobs[id]
	if ok {
		delete(h.blobs, id)
	}
	return b, ok
}

func (h *fileHub) peek(id string) (fileBlob, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	b, ok := h.blobs[id]
	return b, ok
}

// gc entfernt Übertragungen, die älter als 5 Minuten nicht abgeholt wurden.
func (h *fileHub) gc() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		h.mu.Lock()
		for id, b := range h.blobs {
			if time.Since(b.ts) > 5*time.Minute {
				delete(h.blobs, id)
			}
		}
		h.mu.Unlock()
	}
}

// --- Agent-Endpoints (Bearer-Token) ---

// handleAgentFileUpload nimmt die vom Agent gelesene Datei entgegen (Download-Richtung).
func (s *Server) handleAgentFileUpload(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
	if cmd == "" {
		s.writeErr(w, http.StatusBadRequest, "cmd fehlt")
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxFileTransfer+1))
	if err != nil || len(data) > maxFileTransfer {
		s.writeErr(w, http.StatusRequestEntityTooLarge, "Datei zu groß")
		return
	}
	s.files.put(cmd, r.URL.Query().Get("name"), data)
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleAgentFileDownload liefert dem Agent die vom Browser hochgeladenen Bytes (Upload-Richtung).
func (s *Server) handleAgentFileDownload(w http.ResponseWriter, r *http.Request) {
	b, ok := s.files.peek(r.URL.Query().Get("cmd"))
	if !ok {
		s.writeErr(w, http.StatusNotFound, "kein Transfer")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(b.data)
}

// --- Nutzer-Endpoints (Techniker/Admin) ---

// handleBrowse reiht eine Verzeichnisauflistung ein (Ergebnis per Polling).
func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if !s.decodeJSON(w, r, &req) {
		return
	}
	id, err := s.queueCommand(r.Context(), chi.URLParam(r, "id"), "browse_dir", "Dateien auflisten",
		map[string]any{"path": req.Path})
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, map[string]string{"command_id": id})
}

// handleReadFile reiht einen Datei-Download ein (Agent lädt die Datei zum Server).
func (s *Server) handleReadFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if !s.decodeJSON(w, r, &req) {
		return
	}
	if req.Path == "" {
		s.writeErr(w, http.StatusBadRequest, "pfad fehlt")
		return
	}
	id, err := s.queueCommand(r.Context(), chi.URLParam(r, "id"), "read_file", "Datei lesen: "+req.Path,
		map[string]any{"path": req.Path})
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, map[string]string{"command_id": id})
}

// handleServeFile streamt die vom Agent hochgeladene Datei an den Browser.
func (s *Server) handleServeFile(w http.ResponseWriter, r *http.Request) {
	b, ok := s.files.take(chi.URLParam(r, "cmd"))
	if !ok {
		s.writeErr(w, http.StatusNotFound, "Datei nicht (mehr) verfügbar")
		return
	}
	name := b.name
	if name == "" {
		name = "download"
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	_, _ = w.Write(b.data)
}

// handleWriteFile nimmt eine hochgeladene Datei entgegen und reiht das Schreiben ein.
// Body = rohe Datei; Pfad im Query-Parameter „path".
func (s *Server) handleWriteFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeErr(w, http.StatusBadRequest, "path fehlt")
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxFileTransfer+1))
	if err != nil || len(data) > maxFileTransfer {
		s.writeErr(w, http.StatusRequestEntityTooLarge, "Datei zu groß (max 32 MB)")
		return
	}
	xfer := newSessionID()
	s.files.put(xfer, "", data)
	id, err := s.queueCommand(r.Context(), chi.URLParam(r, "id"), "write_file", "Datei schreiben: "+path,
		map[string]any{"path": path, "xfer": xfer})
	if err != nil {
		s.mapStoreErr(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, map[string]string{"command_id": id})
}
