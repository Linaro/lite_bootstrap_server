package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cakeyCmd represents the cakey command
var cakeyCmd = &cobra.Command{
	Use:   "cakey",
	Short: "CA key management",
	Long:  `Management and generation of the certificate authority key.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cakey called")
	},
}

func init() {
	rootCmd.AddCommand(cakeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cakeyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cakeyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
