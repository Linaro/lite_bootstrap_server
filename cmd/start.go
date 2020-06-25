package cmd

import (
	"github.com/microbuilder/linaroca/httpserver"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTPS server and REST API",
	Long: `Starts the HTTPS server, and enables access to CA functions via
a REST API.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := serverCmd.Flags().GetInt16("port")
		httpserver.Start(port)
	},
}

func init() {
	serverCmd.AddCommand(startCmd)
}
