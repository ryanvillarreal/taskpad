package handler

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/service"
)

var reDateParam = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// DailyNoteHandler holds HTTP handlers for daily note endpoints.
type DailyNoteHandler struct {
	svc service.DailyNoteService
}

// NewDailyNoteHandler creates a new DailyNoteHandler.
func NewDailyNoteHandler(svc service.DailyNoteService) *DailyNoteHandler {
	return &DailyNoteHandler{svc: svc}
}

// RegisterRoutes registers daily note routes. Must be called before NoteHandler.RegisterRoutes
// so that /api/v1/notes/daily/{date} takes precedence over /api/v1/notes/{id}.
func (h *DailyNoteHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/notes/daily/{date}", h.get)
	mux.HandleFunc("PUT /api/v1/notes/daily/{date}", h.upsert)
}

func (h *DailyNoteHandler) get(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	if !reDateParam.MatchString(date) {
		respondError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}

	note, err := h.svc.Get(date)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, note)
}

func (h *DailyNoteHandler) upsert(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	if !reDateParam.MatchString(date) {
		respondError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}

	var req model.UpsertDailyNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	note, err := h.svc.Upsert(date, req.Content)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, note)
}
