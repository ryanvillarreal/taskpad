package notes

import (
	"io"
	"net/http"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/server"
)

type Handler struct {
	svc *Service
}

var defaultHandler *Handler

func init() {
	cfg := config.Load()
	defaultHandler = &Handler{svc: NewService(NewStore(cfg.NotesDir))}

	server.Register(
		server.Route{Pattern: "GET /notes", Handler: defaultHandler.count},
		server.Route{Pattern: "POST /notes/{id}", Handler: defaultHandler.save},
		server.Route{Pattern: "GET /notes/{id}", Handler: defaultHandler.get},
		server.Route{Pattern: "DELETE /notes/{id}", Handler: defaultHandler.delete},
	)
}

func (h *Handler) count(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.Count()
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusOK, map[string]int{"count": n})
}

func (h *Handler) save(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	defer r.Body.Close()

	n, err := h.svc.Save(id, string(body))
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusOK, map[string]any{
		"id":         n.ID,
		"created_at": n.CreatedAt,
		"updated_at": n.UpdatedAt,
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	raw, err := h.svc.Raw(id)
	if err == ErrNotFound {
		server.Error(w, http.StatusNotFound, "note not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write(raw)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.svc.Delete(id)
	if err == ErrNotFound {
		server.Error(w, http.StatusNotFound, "note not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
