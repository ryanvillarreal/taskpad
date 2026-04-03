package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/service"
)

// NoteHandler holds HTTP handlers for note endpoints.
type NoteHandler struct {
	svc service.NoteService
}

// NewNoteHandler creates a new NoteHandler.
func NewNoteHandler(svc service.NoteService) *NoteHandler {
	return &NoteHandler{svc: svc}
}

// RegisterRoutes registers all note routes on the given mux.
func (h *NoteHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/notes", h.list)
	mux.HandleFunc("POST /api/v1/notes", h.create)
	mux.HandleFunc("GET /api/v1/notes/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/notes/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/notes/{id}", h.delete)
	mux.HandleFunc("POST /api/v1/notes/bulk/delete", h.bulkDelete)
}

func (h *NoteHandler) list(w http.ResponseWriter, r *http.Request) {
	params := parseListParams(r)
	filters := model.NoteFilters{}

	if v := r.URL.Query().Get("tag"); v != "" {
		filters.Tag = &v
	}
	if v := r.URL.Query().Get("search"); v != "" {
		filters.Search = &v
	}

	result, err := h.svc.List(params, filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list notes")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *NoteHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	note, err := h.svc.Create(req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, note)
}

func (h *NoteHandler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	note, err := h.svc.GetByID(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	var req model.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	note, err := h.svc.Update(id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandler) bulkDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	affected, err := h.svc.BulkDelete(req.IDs)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]int64{"affected": affected})
}
