package cmd

import (
	"github.com/microbuilder/linaroca/caserver"
	"github.com/microbuilder/linaroca/mtlsserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTPS server and REST API",
	Long: `Starts the HTTPS server, and enables access to CA functions via
a REST API.`,
	Run: func(cmd *cobra.Command, args []string) {
		mport := viper.GetInt("server.mport")
		go mtlsserver.StartTCP(int16(mport))
		port := viper.GetInt("server.port")
		caserver.Start(int16(port))
	},
}

func init() {
	serverCmd.AddCommand(startCmd)
}
