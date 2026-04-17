package notes

import (
	"fmt"
	"net/http"

	"github.com/ryanvillarreal/taskpad/internal/server"
)

type Handler struct {
}

// self-register note routes with the server on package import
func init() {
	h := &Handler{}
	server.Register(
		server.Route{Pattern: "GET /note", Handler: h.list},
		server.Route{Pattern: "POST /note", Handler: h.create},
		server.Route{Pattern: "GET /note/{id}", Handler: h.getByID},
	)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Listing all notes")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Creating a note")
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // Available in Go 1.22+
	fmt.Fprintf(w, "Getting note: %s", id)
}
