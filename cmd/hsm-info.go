package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Linaro/lite_bootstrap_server/hsm"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var hsmInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "PKCS#11 Module Info",
	Long:  `Display technical details of the PKCS#11 HSM interface being used.`,
	Run: func(cmd *cobra.Command, args []string) {

		// HSM Info
		if err := hsm.DisplayHSMInfo(); err != nil {
			if errors.Is(err, hsm.ErrQueryFailure) {
				fmt.Println("Error trying to get module info from the PKCS#11 module.")
				os.Exit(1)
			}
			// Catch all handler
			fmt.Printf("Unhandled PKCS#11 error: %s\n", err)
			os.Exit(1)
		}

		// Token/Slot Info
		if err := hsm.DisplaySlotInfo(); err != nil {
			if errors.Is(err, hsm.ErrQueryFailure) {
				fmt.Println("Error trying to get slot 0 info from the PKCS#11 module.")
				os.Exit(1)
			}
			// Catch all handler
			fmt.Printf("Unhandled PKCS#11 error: %s\n", err)
			os.Exit(1)
		}

		// Generate test keys with:
		//
		// $ #ECUUID=$(uuidgen | tr -d -)
		// $ #RSAUUID=$(uuidgen | tr -d -)
		// $ ECUUID=90adf2460d194726ac2235071c5c148b
		// $ RSAUUID=8ce44ddceced463bb6e191efcbb25edb
		// $ pkcs11-tool --module /opt/homebrew/Cellar/softhsm/2.6.1/lib/softhsm/libsofthsm2.so \
		//   -l -p 1234 -k --id $ECUUID --label "Test EC Key" \
		//   --key-type EC:prime256v1
		//
		// $ pkcs11-tool --module /opt/homebrew/Cellar/softhsm/2.6.1/lib/softhsm/libsofthsm2.so \
		//   -l -p 1234 -k --id $RSAUUID --label "Test RSA Key" \
		//   --key-type rsa:2048
		//
		// The keys can be deleted with:
		//
		// $ pkcs11-tool --module <...> \
		//   -l -p 1234 -b --type privkey --label "Test EC Key"
		//
		// $ pkcs11-tool --module <...> \
		//   -l -p 1234 -b --type pubkey --label "Test EC Key"
		//
		// etc for the RSA key

		// Search for the Root CA Key
		// uuid, _ := uuid.Parse("8ce44ddc-eced-463b-b6e1-91efcbb25edb") // RSA
		uuid, _ := uuid.Parse("90adf246-0d19-4726-ac22-35071c5c148b") // EC
		if err := hsm.FindObjectsByUUID(uuid, 10); err != nil {
			fmt.Printf("Unhandled PKCS#11 error: %s\n", err)
			os.Exit(1)
		}

		if err := hsm.FindObjectsByLabel("Test EC Key", 10); err != nil {
			fmt.Printf("Unhandled PKCS#11 error: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	hsmCmd.AddCommand(hsmInfoCmd)
}
