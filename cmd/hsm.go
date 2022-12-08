package cmd

import (
	"github.com/spf13/cobra"
)

var hsmCmd = &cobra.Command{
	Use:   "hsm",
	Short: "PKCS#11 hardware security module key management",
	Long:  `Key management using a PKCS#11 compliant HSM module.`,
}

func init() {
	rootCmd.AddCommand(hsmCmd)
}
