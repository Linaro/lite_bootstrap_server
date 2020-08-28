package httpserver

import (
	"crypto/x509"
	"fmt"
)

// handleCSR processes an incoming CSR, and if valid, builds a
// certificate for the device.
func handleCSR(asn1Data []byte) error {
	csr, err := x509.ParseCertificateRequest(asn1Data)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	fmt.Printf("CSR: %v\n", csr)

	return nil
}
