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

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "list tasks due today",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := tasks.NewService(tasks.NewStore(cfg.TasksDir))
		ts, err := svc.Today(time.Now())
		if err != nil {
			slog.Error("today failed", "err", err)
			os.Exit(1)
		}
		if len(ts) == 0 {
			fmt.Println("no tasks due today")
			return
		}
		for _, t := range ts {
			fmt.Printf("%s  %s  due %s\n", t.ID, t.Title, t.DueAt.Local().Format("3:04pm"))
		}
	},
}

func init() {
	TaskCmd.AddCommand(todayCmd)
}
