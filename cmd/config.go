package cmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/editor"
	"github.com/spf13/cobra"
)

// quick edit the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "open the taskpad config file in $EDITOR",
	Run: func(cmd *cobra.Command, args []string) {
		config.Load()
		path := config.ConfigPath()
		slog.Info("opening config", "path", path)
		if err := editor.Run(path); err != nil {
			slog.Error("editor failed", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
