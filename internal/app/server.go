package app

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rvillarreal/taskpad/internal/config"
	"github.com/rvillarreal/taskpad/internal/database"
	"github.com/rvillarreal/taskpad/internal/handler"
	"github.com/rvillarreal/taskpad/internal/middleware"
	"github.com/rvillarreal/taskpad/internal/repository"
	"github.com/rvillarreal/taskpad/internal/service"
)

func RunServer(cfg config.Config, migrations fs.FS) error {
	db, err := database.Open(cfg.Server.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := database.MigrateFS(db, migrations); err != nil {
		return err
	}

	todoRepo := repository.NewTodoRepository(db)
	noteRepo := repository.NewNoteRepository(db)
	dailyNoteRepo := repository.NewDailyNoteRepository(db)
	todoSvc := service.NewTodoService(todoRepo)
	noteSvc := service.NewNoteService(noteRepo)
	dailyNoteSvc := service.NewDailyNoteService(dailyNoteRepo)
	todoHandler := handler.NewTodoHandler(todoSvc)
	noteHandler := handler.NewNoteHandler(noteSvc)
	dailyNoteHandler := handler.NewDailyNoteHandler(dailyNoteSvc)

	mux := http.NewServeMux()
	todoHandler.RegisterRoutes(mux)
	dailyNoteHandler.RegisterRoutes(mux) // must be before noteHandler to win over /notes/{id}
	noteHandler.RegisterRoutes(mux)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	var h http.Handler = mux
	h = middleware.Auth(cfg.APIKey)(h)
	h = middleware.CORS(cfg.CORSOrigins)(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	log.Println("Server stopped")
	return nil
}
