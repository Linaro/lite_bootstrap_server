package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Linaro/lite_bootstrap_server/hsm"
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
	cobra.OnInitialize(initConfig, initHSM)

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

// InitHSM initialises the hardware secure module using the PKCS#11 interface
func initHSM() {
	pkcs11Module := viper.GetString("hsm.module")
	pkcs11Pin := viper.GetString("hsm.pin")

	// Show a useful warning if no HSM module specified
	if pkcs11Module == "" {
		fmt.Printf("No PKCS#11 module detected.\n\n")
		fmt.Printf("This server relies on a hardware secure module (HSM) being present for\n")
		fmt.Printf("secure key storage, communicating over a standard PKCS#11 interface.\n\n")
		fmt.Printf("If using the default setup, make sure SoftHSM is installed and configured\n")
		fmt.Printf("on your system, and specify the module's path/filename via '--hsm-module',\n")
		fmt.Printf("or via 'module' in the '[hsm]' section of the config file.")
		os.Exit(1)
	}

	// Show a useful warning if no HSM pin specified
	if pkcs11Pin == "" {
		fmt.Printf("No PKCS#11 pin specified.\n\n")
		fmt.Printf("Access to the HSM requires a pin to be specified. Please specify the pin\n")
		fmt.Printf("via '--hsm-pin' or via 'pin' in the '[hsm]' section of the config file.")
		os.Exit(1)
	}

	// Display the HSM module being used
	fmt.Printf("Using PKCS#11 module: %s\n", pkcs11Module)

	// Do some exhaustive error checking to make sure the HSM is setup properly
	err := hsm.TestConnection()
	if err != nil {
		// Incorrect path/filename or init failure
		if errors.Is(err, hsm.ErrInvalidFilename) || errors.Is(err, hsm.ErrInitFailure) {
			fmt.Printf("Invalid PKCS#11 path/filename: %s\n", pkcs11Module)
			fmt.Println("Please set a valid PKCS#11 module path/filename via '--hsm-module'")
			os.Exit(1)
		}

		// No slot(s) defined
		if errors.Is(err, hsm.ErrMissingSlot) {
			fmt.Println("Unable to get the slot list for the PKCS#11 module.")
			fmt.Println("Create an empty token in slot 0, setting an appropriate pin via:")
			fmt.Println("")
			fmt.Println("$ softhsm2-util --init-token --slot 0 --label \"LITEBoot\" --pin 1234")
			os.Exit(1)
		}

		// Failed opening slot 0
		if errors.Is(err, hsm.ErrSessionFailure) {
			fmt.Println("Unable to open a session for the PKCS#11 module in slot 0.")
			os.Exit(1)
		}

		// Invalid PIN code
		if errors.Is(err, hsm.ErrLoginFailure) {
			fmt.Println("Login failed for PKCS#11 module in slot 0.")
			fmt.Println("Please set a valid PKCS#11 pin via '--hsm-pin'")
			os.Exit(1)
		}

		// Catch all handler
		fmt.Printf("Unhandled PKCS#11 error: %s\n", err)
		os.Exit(1)
	}
}
