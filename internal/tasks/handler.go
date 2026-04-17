package tasks

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/notify"
	"github.com/ryanvillarreal/taskpad/internal/server"
)

type Handler struct {
	svc *Service
}

var defaultHandler *Handler

func init() {
	cfg := config.Load()
	svc := NewService(NewStore(cfg.TasksDir))
	defaultHandler = &Handler{svc: svc}

	n := buildNotifier(cfg)
	StartScheduler(svc, n, cfg.Notify.RepeatMinutes)

	server.Register(
		server.Route{Pattern: "GET /tasks", Handler: defaultHandler.list},
		server.Route{Pattern: "POST /tasks", Handler: defaultHandler.create},
		server.Route{Pattern: "GET /tasks/{id}", Handler: defaultHandler.get},
		server.Route{Pattern: "PATCH /tasks/{id}/status", Handler: defaultHandler.setStatus},
		server.Route{Pattern: "DELETE /tasks/{id}", Handler: defaultHandler.delete},
	)
}

func buildNotifier(cfg config.Config) notify.Notifier {
	if cfg.Notify.Backend == "ntfy" && cfg.Notify.Topic != "" {
		slog.Info("notify backend ready", "backend", "ntfy", "topic", cfg.Notify.Topic)
		return notify.NewNtfy(cfg.Notify.URL, cfg.Notify.Topic)
	}
	slog.Info("notify backend not configured, notifications disabled")
	return notify.NullNotifier{}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.svc.List()
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusOK, tasks)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	defer r.Body.Close()

	var req struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Title == "" {
		server.Error(w, http.StatusBadRequest, "title required")
		return
	}

	t, err := h.svc.Create(req.Title)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusCreated, t)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.svc.Get(id)
	if err == ErrNotFound {
		server.Error(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusOK, t)
}

func (h *Handler) setStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	defer r.Body.Close()

	var req struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Status == "" {
		server.Error(w, http.StatusBadRequest, "status required")
		return
	}

	t, err := h.svc.SetStatus(id, Status(req.Status))
	if err == ErrNotFound {
		server.Error(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusOK, t)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.svc.Delete(id)
	if err == ErrNotFound {
		server.Error(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
