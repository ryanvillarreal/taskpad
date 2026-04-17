package taskscmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
	"github.com/spf13/cobra"
)

var createTitle string

var TaskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"t"},
	Short:   "manage tasks and reminders",
	RunE: func(cmd *cobra.Command, args []string) error {
		if createTitle != "" {
			return createTask(createTitle)
		}
		return cmd.Help()
	},
}

func init() {
	TaskCmd.Flags().StringVarP(&createTitle, "create", "c", "", "create a new task")
}

func createTask(title string) error {
	cfg := config.Load()
	svc := tasks.NewService(tasks.NewStore(cfg.TasksDir))
	t, err := svc.Create(title)
	if err != nil {
		slog.Error("create task failed", "err", err)
		os.Exit(1)
	}
	if !t.DueAt.IsZero() {
		fmt.Printf("%s  %s  (due %s)\n", t.ID, t.Title, t.DueAt.Local().Format("Mon Jan 2 3:04pm"))
	} else {
		fmt.Printf("%s  %s\n", t.ID, t.Title)
	}
	return nil
}
