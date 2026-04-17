package taskscmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
	"github.com/spf13/cobra"
)

var dueCmd = &cobra.Command{
	Use:   "due",
	Short: "list overdue and upcoming tasks",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := tasks.NewService(tasks.NewStore(cfg.TasksDir))
		now := time.Now()
		ts, err := svc.Due(now)
		if err != nil {
			slog.Error("due failed", "err", err)
			os.Exit(1)
		}
		if len(ts) == 0 {
			fmt.Println("no overdue tasks")
			return
		}
		for _, t := range ts {
			fmt.Printf("%s  %-40s  due %s\n", t.ID, t.Title, t.DueAt.Local().Format("Mon Jan 2 3:04pm"))
		}
	},
}

func init() {
	TaskCmd.AddCommand(dueCmd)
}
