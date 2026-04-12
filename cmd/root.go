package cmd

import (
	"github.com/rvillarreal/taskpad/internal/config"

	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "taskpad",
	Short: "note taking app for the lols",
}

func Execute() {
	err := rootCmd.Execute()
	// as quickly as possible check args
	if err != nil {
		os.Exit(1)
	}

	cfg = config.Load()
}

// if needed later
// func init() {
// }
