package cmd

import (
	"fmt"
	"log"
	"os"

	// "github.com/Linaro/lite_bootstrap_server/hsm"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "liteboot",
	Short: "Linaro bootstrap and certificate authority server",
	Long:  `A proof-of-concept certificate authority (CA) and management tool.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default \"$HOME/.liteboot.yaml\")")
	rootCmd.PersistentFlags().String("hsm-module", "", "PKCS#11 module path and filename")
	rootCmd.PersistentFlags().String("hsm-pin", "1234", "PKCS#11 module pin")

	viper.BindPFlag("hsm.module", rootCmd.PersistentFlags().Lookup("hsm-module"))
	viper.BindPFlag("hsm.pin", rootCmd.PersistentFlags().Lookup("hsm-pin"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".liteboot" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName(".liteboot")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
