package mtlsserver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
)

func handleConnection(c net.Conn) {
	fmt.Println("Connection accepted from", c.RemoteAddr())

	// Get TLS connection
	tlsConn, ok := c.(*tls.Conn)
	if !ok {
		fmt.Println("Not a TLS connection!")
		return
	}

	// Start the TLS handshake process
	// This will also validate the client cert via validatePeer
	if err := tlsConn.Handshake(); err != nil {
		fmt.Println("Client handshake error:", err)
		return
	}

	// Get client certificate using TLS ConnectionState
	state := tlsConn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Printf("Client certificate:\n")
		fmt.Printf("- Issuer CN: %s\n", v.Issuer.CommonName)
		fmt.Printf("- Subject: %s\n", v.Subject)
	}

	// Close connection
	c.Close()
}

// Starts a TCP server with mTLS authentication
func StartTCP(port int16) {
	// Create a certificate pool with the CA certificate
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("certs/CA.crt")
	if err != nil {
		log.Fatal(err)
	}
	certPool.AppendCertsFromPEM(caCert)

	// Load server key pair
	cer, err := tls.LoadX509KeyPair("certs/SERVER.crt", "certs/SERVER.key")
	if err != nil {
		log.Fatal(err)
	}

	// Construct a TLS config with our CA and server certificates
	config := tls.Config{
		// Set the minimum TLS version to 1.2
		MinVersion: tls.VersionTLS12,
		// Set the CA certificate to verify client certs against
		ClientCAs: certPool,
		// Set the server certificate
		Certificates: []tls.Certificate{cer},
		// Alt: RequestClientCert
		ClientAuth: tls.RequireAndVerifyClientCert,
		// Callback to verify client cert details
		VerifyPeerCertificate: validatePeer,
	}

	// Listen for TCP connections
	fmt.Println("Starting mTLS TCP server on localhost:" + strconv.Itoa(int(port)))
	listener, err := tls.Listen("tcp", "localhost:"+strconv.Itoa(int(port)), &config)
	if err != nil {
		fmt.Println("Unable to start listening")
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Unable to accept incoming connection, error:", err)
			continue
		}

		// Concurrent connection handling
		go handleConnection(conn)
	}
}

// ValidatePeer checks the given certificates and makes sure they are
// appropriate for requests to this TCP server
func validatePeer(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(verifiedChains) != 1 {
		return fmt.Errorf("expecting a single certificate chain")
	}

	// TODO: Validate client certificate UUID is valid in cadb
	// log.Printf("cert: %#v", verifiedChains[0][0].Subject)

	return nil
}
