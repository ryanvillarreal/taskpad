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

/*
exposed:
Execute() - main entry point into Cobra

init() - start logging first
*/
func init() {
	// start logging before anything else
	logs.Start()
}

func Execute() {
	err := rootCmd.Execute()

	// now that logging is on we have visibility
	// Check for verbose flag to increase logging
	if err != nil {
		slog.Error("Ruh Roh Raggy:", err)
		os.Exit(1)
	}

	cfg := config.Load()
	slog.Info("%s", cfg)
}
