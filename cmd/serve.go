package cmd

import (
	_ "github.com/ryanvillarreal/taskpad/internal/links" // registers link routes
	_ "github.com/ryanvillarreal/taskpad/internal/notes" // registers note routes
	_ "github.com/ryanvillarreal/taskpad/internal/tasks" // registers task routes + scheduler
	"github.com/ryanvillarreal/taskpad/internal/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the API server ",
	Run: func(cmd *cobra.Command, args []string) {
		server.RunServer()
	},
}

// use the init func to register itself
func init() {
	rootCmd.AddCommand(serveCmd)
}
