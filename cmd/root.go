package cmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	logs "github.com/ryanvillarreal/taskpad/internal/logging"
	notescmd "github.com/ryanvillarreal/taskpad/cmd/notes"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "taskpad",
	Short: "note taking app for the lols",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logs.Start(verbose)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
	rootCmd.AddCommand(notescmd.NoteCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Ruh Roh Raggy", "err", err)
		os.Exit(1)
	}

	cfg := config.Load()
	slog.Debug("Config loaded", "cfg", cfg)
}
