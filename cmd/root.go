package cmd

import (
	"log/slog"
	"os"

	"github.com/ryanvillarreal/taskpad/internal/config"
	logs "github.com/ryanvillarreal/taskpad/internal/logging"
	linkscmd "github.com/ryanvillarreal/taskpad/cmd/links"
	notescmd "github.com/ryanvillarreal/taskpad/cmd/notes"
	searchcmd "github.com/ryanvillarreal/taskpad/cmd/search"
	taskscmd "github.com/ryanvillarreal/taskpad/cmd/tasks"
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
	rootCmd.AddCommand(taskscmd.TaskCmd)
	rootCmd.AddCommand(linkscmd.LinkCmd)
	rootCmd.AddCommand(searchcmd.SearchCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Ruh Roh Raggy", "err", err)
		os.Exit(1)
	}

	cfg := config.Load()
	slog.Debug("Config loaded", "cfg", cfg)
}
