package notescmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/notes"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list all notes (newest first)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := notes.NewService(notes.NewStore(cfg.NotesDir))
		ids, err := svc.List()
		if err != nil {
			slog.Error("list failed", "err", err)
			os.Exit(1)
		}
		slog.Debug("listed notes", "count", len(ids), "dir", cfg.NotesDir)
		for _, id := range ids {
			fmt.Println(id)
		}
	},
}

func init() {
	NoteCmd.AddCommand(lsCmd)
}
