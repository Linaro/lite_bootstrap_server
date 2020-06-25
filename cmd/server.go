package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "HTTPS server management",
	Long: `Starts the HTTPS server and REST API that can be used to sign new
certificate signing requests (CSRs), and verify previous device registrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("server called")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Allow a custom port number
	serverCmd.PersistentFlags().Int16P("port", "p", 443, "Port number")
}
