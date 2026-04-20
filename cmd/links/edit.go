package linkscmd

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/editor"
	"github.com/ryanvillarreal/taskpad/internal/links"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "open a link in $EDITOR",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := links.NewService(links.NewStore(cfg.LinksDir), cfg.FetchLimit)
		id, err := svc.Resolve(args[0])
		if err != nil {
			slog.Error("edit failed", "prefix", args[0], "err", err)
			os.Exit(1)
		}
		path := filepath.Join(cfg.LinksDir, id+".md")
		if err := editor.Run(path); err != nil {
			slog.Error("editor failed", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	LinkCmd.AddCommand(editCmd)
}
