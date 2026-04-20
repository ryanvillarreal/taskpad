package linkscmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/links"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list saved links",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		svc := links.NewService(links.NewStore(cfg.LinksDir), cfg.FetchLimit)
		all, err := svc.List()
		if err != nil {
			slog.Error("list links failed", "err", err)
			os.Exit(1)
		}
		for _, l := range all {
			fetched := ""
			if l.Fetched {
				fetched = " [fetched]"
			}
			fmt.Printf("%s  %-50s  %s%s\n", l.ID, l.Title, l.URL, fetched)
		}
	},
}

func init() {
	LinkCmd.AddCommand(lsCmd)
}
