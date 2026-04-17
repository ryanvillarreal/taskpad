package notescmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/client"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "delete a note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := client.DeleteNote(args[0]); err != nil {
			slog.Error("rm failed", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	NoteCmd.AddCommand(rmCmd)
}
