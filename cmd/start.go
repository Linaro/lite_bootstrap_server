package cmd

import (
	"os"

	"github.com/Linaro/lite_bootstrap_server/caserver"
	"github.com/Linaro/lite_bootstrap_server/mtlsserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Gets the hostname from the command line, the config file, the CAHOSTNAME
// env variable, the OS, or default to 'localhost' in that priority
func getHostname() string {
	// 1. '--hostname' from cli args or 'hostname' from .liteboot.toml
	// 2. 'hostname' from .liteboot.toml (lower priority if 1, 2 both defined)
	hostname := viper.GetString("server.hostname")
	if hostname != "" {
		return hostname
	}

	// 3. '$CAHOSTNAME' environment variable
	hostname = os.Getenv("CAHOSTNAME")
	if hostname != "" {
		return hostname
	}

	// 4. os.Hostname() (bash $HOSTNAME, zsh $HOST, etc.)
	var err error
	hostname, err = os.Hostname()
	if (err != nil) || (hostname == "") {
		// 5. As a last resort, fall back to localhost
		hostname = "localhost"
	}

	return hostname
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTPS server and REST API",
	Long: `Starts the HTTPS server, and enables access to CA functions via
a REST API.`,
	Run: func(cmd *cobra.Command, args []string) {
		hostname := getHostname()
		mport := viper.GetInt("server.mport")
		go mtlsserver.StartTCP(hostname, int16(mport))
		port := viper.GetInt("server.port")
		caserver.Start(hostname, int16(port))
	},
}

func init() {
	serverCmd.AddCommand(startCmd)
}
