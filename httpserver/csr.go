package httpserver

import (
	"crypto/x509"
	"fmt"
	"time"

	"github.com/microbuilder/linaroca/cadb"
	"github.com/microbuilder/linaroca/signer"
)

// handleCSR processes an incoming CSR, and if valid, builds a
// certificate for the device.
func handleCSR(asn1Data []byte) ([]byte, error) {
	csr, err := x509.ParseCertificateRequest(asn1Data)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil, err
	}

	// TODO: Need to validate all of the information from the
	// certificate request.

	db, err := cadb.Open()
	if err != nil {
		return nil, err
	}

	ser, err := db.GetSerial()

	cert := &x509.Certificate{
		SerialNumber: ser,
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		// TODO: Extensions that make sense to us.
	}

	// fmt.Printf("CSR: %v\n", csr)
	// fmt.Printf("Cert: %v\n", cert)

	signedCert, err := signCert(cert)
	if err != nil {
		fmt.Printf("Sign error: %v\n", err)
		return nil, err
	}

	id := cert.Subject.CommonName
	err = db.AddCert(id, ser, signedCert)
	if err != nil {
		fmt.Printf("Add cert err: %v\n", err)
		return nil, err
	}

	return signedCert, nil
}

func signCert(template *x509.Certificate) ([]byte, error) {
	// TODO: Don't use hardcoded names here.
	// TODO: This can probably share a bit of code with the root
	// cert generation.

	sig, err := signer.LoadSigningCert("CA")
	if err != nil {
		return nil, err
	}

	cert, err := sig.SignTemplate(template)
	if err != nil {
		return nil, err
	}

	// TODO: Put cert into database

	// fmt.Printf("New cert: %v\n", cert)

	return cert, nil
}
