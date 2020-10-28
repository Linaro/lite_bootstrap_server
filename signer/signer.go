// Package signer combines some utilities for management behind the
// keys involved in certificates.
package signer // "github.com/microbuilder/linaroca/signer"

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

// A SigningCert is a certificate that can be used to sign other
// certificates.  As such, it's private key is also needed.
type SigningCert struct {
	CertBin    []byte
	Cert       *x509.Certificate
	PrivateKey *ecdsa.PrivateKey
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
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		KeyUsage:              x509.KeyUsageCertSign,
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Determine the KeyID based on the sha1 of the marshalled
	// public key.
	pubBytes := elliptic.Marshal(privKey.Curve, privKey.X, privKey.Y)
	keyId := sha1.Sum(pubBytes)

	ca.SubjectKeyId = keyId[:]
	ca.AuthorityKeyId = keyId[:]

	// Self sign this key.
	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, err
	}

	return &SigningCert{
		CertBin:    cert,
		PrivateKey: privKey,
	}, nil
}

func (s *SigningCert) SignTemplate(template *x509.Certificate, pub interface{}) ([]byte, error) {
	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Expecting ECDSA key on CSR")
	}

	// Fill in the SubjectKeyId in the template, based on the
	// public key.  The AuthorityKeyId will be filled in by the
	// x509 library, provided we're using a recent enough version
	// of Go.
	pubBytes := elliptic.Marshal(ecPub.Curve, ecPub.X, ecPub.Y)
	keyId := sha1.Sum(pubBytes)

	template.SubjectKeyId = keyId[:]

	return x509.CreateCertificate(rand.Reader, template, s.Cert,
		pub, s.PrivateKey)
}

// LoadSigningCert loads a signing certificate from a pair of files
// base.crt, and base.key.
func LoadSigningCert(base string) (*SigningCert, error) {
	// TODO: Load cert and key

	certBin, err := loadPem(base+".crt", "CERTIFICATE")
	if err != nil {
		return nil, err
	}

	caCert, err := x509.ParseCertificate(certBin)
	if err != nil {
		return nil, err
	}

	keyBin, err := loadPem(base+".key", "EC PRIVATE KEY")
	if err != nil {
		return nil, err
	}

	key, err := x509.ParseECPrivateKey(keyBin)
	if err != nil {
		return nil, err
	}

	return &SigningCert{
		CertBin:    certBin,
		Cert:       caCert,
		PrivateKey: key,
	}, nil
}

// loadPem loads a file of an expected type in PEM form.
func loadPem(name, expectedType string) ([]byte, error) {
	pems, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	bin, rest := pem.Decode(pems)
	if err != nil {
		return nil, err
	}
	if bin.Type != expectedType {
		return nil, fmt.Errorf("Expecting BEGIN %s", expectedType)
	}
	if len(rest) != 0 {
		return nil, errors.New("Extraneous file data after certificate")
	}

	return bin.Bytes, nil
}

// Export writes the certificate and the signing key to files in PEM
// format.  May return an error if the files exist.
func (s *SigningCert) Export(cafile, keyfile string) error {
	err := pemWrite(cafile, "CERTIFICATE", s.CertBin)
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
