package cmd

import (
	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/logging"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "taskpad",
	Short: "note taking app for the lols",
}

func Execute() {
	err := rootCmd.Execute()

	// start logging before anything else
	logs.Start()

	// now that logging is on we have visibility
	// Check for verbose flag to increase logging
	if err != nil {
		slog.Error("Ruh Roh Raggy:", err)
		os.Exit(1)
	}

	cfg := config.Load()
	slog.Info("%s", cfg)
}
