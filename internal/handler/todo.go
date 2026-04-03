package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/service"
)

// TodoHandler holds HTTP handlers for todo endpoints.
type TodoHandler struct {
	svc service.TodoService
}

// NewTodoHandler creates a new TodoHandler.
func NewTodoHandler(svc service.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

// RegisterRoutes registers all todo routes on the given mux.
func (h *TodoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/todos", h.list)
	mux.HandleFunc("POST /api/v1/todos", h.create)
	mux.HandleFunc("GET /api/v1/todos/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/todos/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/todos/{id}", h.delete)
	mux.HandleFunc("PATCH /api/v1/todos/{id}/complete", h.setCompleted)
	mux.HandleFunc("POST /api/v1/todos/bulk/complete", h.bulkComplete)
	mux.HandleFunc("POST /api/v1/todos/bulk/delete", h.bulkDelete)
}

func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	params := parseListParams(r)
	filters := model.TodoFilters{}

	if v := r.URL.Query().Get("completed"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			respondError(w, http.StatusBadRequest, "completed must be true or false")
			return
		}
		filters.Completed = &b
	}
	if v := r.URL.Query().Get("priority"); v != "" {
		p := model.Priority(v)
		filters.Priority = &p
	}

	result, err := h.svc.List(params, filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list todos")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	todo, err := h.svc.Create(req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	todo, err := h.svc.GetByID(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	var req model.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	todo, err := h.svc.Update(id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
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

func (h *TodoHandler) setCompleted(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isValidUUID(id) {
		respondError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	var req struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	todo, err := h.svc.SetCompleted(id, req.Completed)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) bulkComplete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	affected, err := h.svc.BulkComplete(req.IDs)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]int64{"affected": affected})
}

func (h *TodoHandler) bulkDelete(w http.ResponseWriter, r *http.Request) {
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

// --- Helpers ---

func parseListParams(r *http.Request) model.ListParams {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	return model.ListParams{
		Limit:  limit,
		Offset: offset,
		Sort:   q.Get("sort"),
		Order:  q.Get("order"),
	}
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "resource not found")
		return
	}
	if errors.Is(err, service.ErrValidation) {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondError(w, http.StatusInternalServerError, "internal server error")
}
