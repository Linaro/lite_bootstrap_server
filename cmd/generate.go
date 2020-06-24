package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

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

		ca, err := NewSigningCert()
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

// A SigningCert is a certificate that can be used to sign other
// certificates.  As such, it's private key is also needed.
type SigningCert struct {
	Certificate []byte
	PrivateKey  *ecdsa.PrivateKey
}

// NewSigningCert builds a fresh signing certificate to use as a root
// certificate.
func NewSigningCert() (*SigningCert, error) {
	ca := &x509.Certificate{
		// TODO: We need to somewhat manage these serial
		// numbers.  Generating from date/time might work.
		// Also, the common name will need to be unique.
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"Linaro, LTD"},
			CommonName:   "LinaroCA Root Cert - 2020",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Self sign this key.
	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, err
	}

	return &SigningCert{
		Certificate: cert,
		PrivateKey:  privKey,
	}, nil
}

// Export writes the certificate and the signing key to files in PEM
// format.  May return an error if the files exist.
func (s *SigningCert) Export(cafile, keyfile string) error {
	err := pemWrite(cafile, "CERTIFICATE", s.Certificate)
	if err != nil {
		return err
	}

	priv, err := x509.MarshalECPrivateKey(s.PrivateKey)
	if err != nil {
		return err
	}

	err = pemWrite(keyfile, "EC PRIVATE KEY", priv)
	if err != nil {
		return err
	}

	return nil
}

// pemWrite writes the given block of data to a pem file of the given
// kind.
func pemWrite(path string, kind string, data []byte) error {
	var buf bytes.Buffer
	pem.Encode(&buf, &pem.Block{
		Type:  kind,
		Bytes: data,
	})

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = fd.Write(buf.Bytes())
	return err
}
