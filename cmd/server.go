package cmd

import (
	"github.com/microbuilder/linaroca/cloud"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "HTTPS server management",
	Long: `Starts the HTTPS server and REST API that can be used to sign new
certificate signing requests (CSRs), and verify previous device registrations.`,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Allow a custom port number
	serverCmd.PersistentFlags().Int16P("port", "p", 1443, "CA port number")
	serverCmd.PersistentFlags().Int16P("mport", "m", 8443, "mTLS port number")

	// Configure the cloud service.
	serverCmd.PersistentFlags().StringVar(&cloud.HubName, "hubname", "hubname", "Azure Hub Name")
	serverCmd.PersistentFlags().StringVar(&cloud.ResourceGroup, "resourcegroup", "resourcegroup", "Azure Resource Group")
}
