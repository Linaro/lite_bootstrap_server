package cmd

import (
	"fmt"
	"path"

	"github.com/microbuilder/linaroca/signer"
	"github.com/spf13/cobra"
)

var cafile = "CA.crt"

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new CA keypair",
	Long: `This command can be used to generate a new keypair for the CA,
which will be used when signing any incoming CSRs. The public key of this
keypair can be used on the device side to verify certificate signatures.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generate called")
		fmt.Printf("use cafile: %s\n", cafile)

		if path.Ext(cafile) != ".crt" {
			fmt.Printf("Expect certificate to end in '.crt'")
			return
		}

		ca, err := signer.NewSigningCert()
		if err != nil {
			fmt.Printf("Unable to create certificate: %s\n", err)
			return
		}

		err = ca.Export(cafile, cafile[:len(cafile)-4]+".key")
		if err != nil {
			fmt.Printf("Unable to write cert: %s\n", err)
			return
		}
	},
}

func init() {
	cakeyCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&cafile, "cafile", cafile, "Filename for generated certificate")
}
