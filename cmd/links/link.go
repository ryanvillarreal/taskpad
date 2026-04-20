package linkscmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/links"
	"github.com/spf13/cobra"
)

var (
	saveURL   string
	noFetch   bool
)

var LinkCmd = &cobra.Command{
	Use:     "link",
	Aliases: []string{"l"},
	Short:   "save and manage links",
	RunE: func(cmd *cobra.Command, args []string) error {
		if saveURL != "" {
			return saveLink(saveURL, !noFetch)
		}
		return cmd.Help()
	},
}

func init() {
	LinkCmd.Flags().StringVarP(&saveURL, "url", "u", "", "url to save")
	LinkCmd.Flags().BoolVar(&noFetch, "no-fetch", false, "skip fetching article content")
}

func saveLink(rawURL string, fetch bool) error {
	cfg := config.Load()
	svc := links.NewService(links.NewStore(cfg.LinksDir), cfg.FetchLimit)
	l, err := svc.Create(rawURL, fetch)
	if err != nil {
		slog.Error("save link failed", "err", err)
		os.Exit(1)
	}
	fmt.Printf("%s  %s\n", l.ID, l.Title)
	if l.Description != "" {
		fmt.Printf("        %s\n", l.Description)
	}
	return nil
}
