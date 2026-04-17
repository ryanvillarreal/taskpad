package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryanvillarreal/taskpad/internal/config"
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
	cfg := config.Load()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("GET /health")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	for _, r := range routes {
		slog.Debug("registering route", "pattern", r.Pattern)
		mux.HandleFunc(r.Pattern, r.Handler)
	}

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	scheme := "http"
	if cfg.TLS.Enabled {
		if cfg.TLS.CertFile == "" || cfg.TLS.KeyFile == "" {
			return fmt.Errorf("tls enabled but cert_file or key_file is empty")
		}
		if _, err := os.Stat(cfg.TLS.CertFile); err != nil {
			return fmt.Errorf("tls cert_file: %w", err)
		}
		if _, err := os.Stat(cfg.TLS.KeyFile); err != nil {
			return fmt.Errorf("tls key_file: %w", err)
		}
		scheme = "https"
	}

	slog.Info("Starting API Server", "scheme", scheme, "addr", addr)

	go func() {
		var err error
		if cfg.TLS.Enabled {
			err = srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
		}
	}()

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
