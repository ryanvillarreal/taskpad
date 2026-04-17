package taskscmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [active|paused|closed|all]",
	Short: "list tasks (default: active)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filter := tasks.StatusActive
		showAll := false
		if len(args) == 1 {
			if args[0] == "all" {
				showAll = true
			} else {
				filter = tasks.Status(args[0])
			}
		}

		cfg := config.Load()
		svc := tasks.NewService(tasks.NewStore(cfg.TasksDir))
		all, err := svc.List()
		if err != nil {
			slog.Error("list tasks failed", "err", err)
			os.Exit(1)
		}
		for _, t := range all {
			if !showAll && t.Status != filter {
				continue
			}
			if t.DueAt.IsZero() {
				fmt.Printf("%s  [%s]  %s\n", t.ID, t.Status, t.Title)
			} else {
				fmt.Printf("%s  [%s]  %-40s  due %s\n", t.ID, t.Status, t.Title, t.DueAt.Local().Format("Mon Jan 2 3:04pm"))
			}
		}
	},
}

func init() {
	TaskCmd.AddCommand(lsCmd)
}
