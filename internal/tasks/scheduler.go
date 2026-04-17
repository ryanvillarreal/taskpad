package tasks

import (
	"log/slog"
	"time"

	"github.com/ryanvillarreal/taskpad/internal/notify"
)

func StartScheduler(svc *Service, n notify.Notifier, repeatMinutes int) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			runNotifications(svc, n, repeatMinutes)
		}
	}()
}

func runNotifications(svc *Service, n notify.Notifier, repeatMinutes int) {
	now := time.Now().UTC()
	due, err := svc.Due(now)
	if err != nil {
		slog.Error("scheduler: list due tasks", "err", err)
		return
	}
	repeat := time.Duration(repeatMinutes) * time.Minute
	for _, t := range due {
		if !t.NotifiedAt.IsZero() && now.Sub(t.NotifiedAt) < repeat {
			continue
		}
		if err := n.Send("taskpad", t.Title); err != nil {
			slog.Error("scheduler: notify failed", "id", t.ID, "err", err)
			continue
		}
		if _, err := svc.MarkNotified(t.ID, now); err != nil {
			slog.Error("scheduler: mark notified", "id", t.ID, "err", err)
		}
		slog.Info("task notified", "id", t.ID, "title", t.Title)
	}
}
