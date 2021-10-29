package mtlsserver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"io"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, world!\n")
}

func Start(port int16) {
	// Set up test resource handler(s)
	http.HandleFunc("/hello", helloHandler)

	// Create a certificate pool with the CA certificate
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("CA.crt")
	if err != nil {
		log.Fatal(err)
	}
	certPool.AppendCertsFromPEM(caCert)

	// Enable TLS client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs:  certPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	// Create a server instance to listen on the appropriatee port
	server := &http.Server{
		Addr:      ":" + strconv.Itoa(int(port)),
		TLSConfig: tlsConfig,
	}

	// Listen for HTTPS connections, using the server certificate
	fmt.Println("Starting mTLS echo server on port https://localhost:" + strconv.Itoa(int(port)))
	log.Fatal(server.ListenAndServeTLS("SERVER.crt", "SERVER.key"))
}
