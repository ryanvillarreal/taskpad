package linkscmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/links"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "delete a saved link",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := links.NewService(links.NewStore(cfg.LinksDir), cfg.FetchLimit)
		id, err := svc.Resolve(args[0])
		if err != nil {
			slog.Error("rm failed", "prefix", args[0], "err", err)
			os.Exit(1)
		}
		if err := svc.Delete(id); err != nil {
			slog.Error("rm failed", "id", id, "err", err)
			os.Exit(1)
		}
		slog.Info("link deleted", "id", id)
	},
}

func init() {
	LinkCmd.AddCommand(rmCmd)
}
