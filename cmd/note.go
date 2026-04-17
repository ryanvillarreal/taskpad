package cmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/client"
	"github.com/spf13/cobra"
)

var noteNew bool

var noteCmd = &cobra.Command{
	Use:   "note [id]",
	Short: "open a note in $EDITOR and sync on save (defaults to today)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := ""
		allowCreate := true
		if len(args) == 1 {
			id = args[0]
			allowCreate = noteNew
		}
		if err := client.EditNote(id, allowCreate); err != nil {
			slog.Error("note failed", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	noteCmd.Flags().BoolVar(&noteNew, "new", false, "create the note if it doesn't exist")
	rootCmd.AddCommand(noteCmd)
}
