package links

import (
	"encoding/json"
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
	defaultHandler = &Handler{svc: NewService(NewStore(cfg.LinksDir), cfg.FetchLimit)}

	server.Register(
		server.Route{Pattern: "GET /links", Handler: defaultHandler.list},
		server.Route{Pattern: "POST /links", Handler: defaultHandler.create},
		server.Route{Pattern: "GET /links/{id}", Handler: defaultHandler.get},
		server.Route{Pattern: "DELETE /links/{id}", Handler: defaultHandler.delete},
	)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	links, err := h.svc.List()
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusOK, links)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		server.Error(w, http.StatusBadRequest, "read body: "+err.Error())
		return
	}
	defer r.Body.Close()

	var req struct {
		URL   string `json:"url"`
		Fetch bool   `json:"fetch"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.URL == "" {
		server.Error(w, http.StatusBadRequest, "url required")
		return
	}

	l, err := h.svc.Create(req.URL, req.Fetch)
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusCreated, l)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	l, err := h.svc.Get(id)
	if err == ErrNotFound {
		server.Error(w, http.StatusNotFound, "link not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	server.JSON(w, http.StatusOK, l)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.svc.Delete(id)
	if err == ErrNotFound {
		server.Error(w, http.StatusNotFound, "link not found")
		return
	}
	if err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
