package taskscmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause <id>",
	Short: "pause a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := tasks.NewService(tasks.NewStore(cfg.TasksDir))
		id, err := svc.Resolve(args[0])
		if err != nil {
			slog.Error("pause failed", "prefix", args[0], "err", err)
			os.Exit(1)
		}
		if _, err := svc.SetStatus(id, tasks.StatusPaused); err != nil {
			slog.Error("pause failed", "id", id, "err", err)
			os.Exit(1)
		}
		slog.Info("task paused", "id", id)
	},
}

func init() {
	TaskCmd.AddCommand(pauseCmd)
}
