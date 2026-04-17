package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
exposed:
RunServer() -
Register() - called from handler packages' init() to add routes
TODO StopServer()
TODO PauseServer()
TODO RestartServer()
*/

// Route describes a single HTTP route to be mounted on the server mux.
type Route struct {
	Pattern string // e.g. "GET /note/{id}"
	Handler http.HandlerFunc
}

// routes is the package-level registry populated by Register.
var routes []Route

// Register adds routes to the server. Call from init() in handler packages.
func Register(rs ...Route) {
	routes = append(routes, rs...)
}

func RunServer() error {
	slog.Info("Starting API Server")

	// create http mux
	mux := http.NewServeMux()

	// create health endpoint - quick and easy
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("GET /health")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// mount routes registered by handler packages
	for _, r := range routes {
		slog.Debug("registering route", "pattern", r.Pattern)
		mux.HandleFunc(r.Pattern, r.Handler)
	}

	var h http.Handler = mux

	//TODO is this taking in account of our config?
	// TODO do we need to handle better defaults here?
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// finally fork into new go proc
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed: %s", err)
		}
	}()

	// setup binds for ctrl + c
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	slog.Info("Server stopped")
	return nil

}
