package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getpubCmd represents the getpub command
var getpubCmd = &cobra.Command{
	Use:   "getpub",
	Short: "Get public CA key",
	Long:  `Returns the public CA key in DER and PEM format.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("getpub called")
	},
}

func init() {
	cakeyCmd.AddCommand(getpubCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getpubCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getpubCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
