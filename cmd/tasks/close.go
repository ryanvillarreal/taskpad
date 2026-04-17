package taskscmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "mark a task as closed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := tasks.NewService(tasks.NewStore(cfg.TasksDir))
		id, err := svc.Resolve(args[0])
		if err != nil {
			slog.Error("close failed", "prefix", args[0], "err", err)
			os.Exit(1)
		}
		if _, err := svc.SetStatus(id, tasks.StatusClosed); err != nil {
			slog.Error("close failed", "id", id, "err", err)
			os.Exit(1)
		}
		slog.Info("task closed", "id", id)
	},
}

func init() {
	TaskCmd.AddCommand(closeCmd)
}
