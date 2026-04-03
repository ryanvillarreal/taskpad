package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rvillarreal/taskpad/internal/database"
	"github.com/rvillarreal/taskpad/internal/handler"
	"github.com/rvillarreal/taskpad/internal/middleware"
	"github.com/rvillarreal/taskpad/internal/repository"
	"github.com/rvillarreal/taskpad/internal/service"
)

func main() {
	// Config from environment with defaults.
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./taskpad.db")
	migrationsDir := getEnv("MIGRATIONS_DIR", "./migrations")

	// Open database.
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations.
	if err := database.Migrate(db, migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Wire up layers.
	todoRepo := repository.NewTodoRepository(db)
	noteRepo := repository.NewNoteRepository(db)

	todoSvc := service.NewTodoService(todoRepo)
	noteSvc := service.NewNoteService(noteRepo)

	todoHandler := handler.NewTodoHandler(todoSvc)
	noteHandler := handler.NewNoteHandler(noteSvc)

	// Register routes.
	mux := http.NewServeMux()
	todoHandler.RegisterRoutes(mux)
	noteHandler.RegisterRoutes(mux)

	// Health check.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply middleware stack: recovery -> logging -> CORS -> routes.
	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	// Create server.
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("Server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown: wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
