package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	serverCmd.PersistentFlags().String("hubname", "hubname", "Azure Hub Name")
	serverCmd.PersistentFlags().String("resourcegroup", "resourcegroup", "Azure Resource Group")
	serverCmd.PersistentFlags().String("mqttport", "mqttport", "Azure MQTT Port")

	viper.BindPFlag("server.hubname", serverCmd.PersistentFlags().Lookup("hubname"))
	viper.BindPFlag("server.resourcegroup", serverCmd.PersistentFlags().Lookup("resourcegroup"))
	viper.BindPFlag("server.port", serverCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("server.mport", serverCmd.PersistentFlags().Lookup("mport"))
	viper.BindPFlag("server.mqttport", serverCmd.PersistentFlags().Lookup("mqttport"))
}
