package caserver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/Linaro/lite_bootstrap_server/cadb"
	"github.com/Linaro/lite_bootstrap_server/protocol"
	"github.com/spf13/viper"
)

// The database itself.
// TODO: Should this be a context instead of a global?
var db *cadb.Conn

// Initialisation request handler
func irPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ir POST called"}
{"serial": "Device serial number"}`))
}

// Certification request handler
func crPost(w http.ResponseWriter, r *http.Request) {
	use_cbor := false
	switch r.Header.Get("Content-Type") {
	case "application/cbor":
		use_cbor = true
	case "application/json":
	case "":
		// Default to JSON if not Content-Type provided (curl, etc.)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request: Content-Type must be application/cbor or application/json"}`))
		return
	}

	var err error
	var req protocol.CSRRequest
	if use_cbor {
		w.Header().Set("Content-Type", "application/cbor")
		dec := cbor.NewDecoder(r.Body)
		err = dec.Decode(&req)
	} else {
		w.Header().Set("Content-Type", "application/json")
		dec := json.NewDecoder(r.Body)
		err = dec.Decode(&req)
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request: POST data did not match specified Content-Type"}`))
		return
	}

	// fmt.Printf("Got csr: %v\n", &req)

	cert, err := handleCSR(req.CSR)
	if err != nil {
		// TODO: Encode the error.
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "err: %v", err)
		return
	}

	if use_cbor {
		w.Header().Set("Content-Type", "application/cbor")
		w.WriteHeader(http.StatusOK)
		enc := cbor.NewEncoder(w)
		err = enc.Encode(&protocol.CSRResponse{
			Status: 0,
			Cert:   cert,
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		err = enc.Encode(&protocol.CSRResponse{
			Status: 0,
			Cert:   cert,
		})
	}
}

// Maximum file size for uploaded CSRs = 4 KB
const MAX_CSR_UPLOAD_SIZE = 1024 * 4

// Certification request from PKCS#10 handler
func p10crPost(w http.ResponseWriter, r *http.Request) {
	// Setup header for errors in JSON format
	w.Header().Set("Content-Type", "application/json")

	// Expect multipart transfer
	err := r.ParseMultipartForm(MAX_CSR_UPLOAD_SIZE)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request (no multipart form)"}`))
		return
	}

	// Validate posted file
	file, fileHeader, err := r.FormFile("csrfile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Missing csrfile file"}`))
		return
	}
	defer file.Close()

	// Check file size
	fileSize := fileHeader.Size
	fmt.Printf("Received CSR file: %v bytes\n", fileSize)
	if fileSize > MAX_CSR_UPLOAD_SIZE {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "File too large"}`))
		return
	}

	// Make sure we can read the file
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "File unreadable"}`))
		return
	}

	// Check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "text/plain; charset=utf-8":
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid file type (require 'text/plain; charset=utf-8')"}`))
		return
	}

	// fmt.Printf("Got csr: \n%s\n", fileBytes)

	// Input file is in PEM format. The payload must be extracted
	// and converted to a binary array, similar to a DER file, before
	// passing it in to handleCSR.
	pemin, rest := pem.Decode(fileBytes)
	if len(rest) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid PEM input. Expecting one block"}`))
		return
	}
	if pemin.Type != "CERTIFICATE REQUEST" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Expecting BEGIN CERTIFICATE REQUEST"}`))
		return
	}

	// Process the CSR and register the certificate details
	cert, err := handleCSR(pemin.Bytes)
	if err != nil {
		// TODO: Encode the error.
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid CSR"}`))
		return
	}

	// Convert DER output to PEM
	pemout := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})

	// Set the file response details
	// MIME type = application/x-x509-user-cert or application/x-pem-file ?
	w.Header().Set("Content-Disposition", "attachment; filename=USERx.der")
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Length", string(len(pemout)))
	io.Copy(w, bytes.NewReader(pemout))
}

// Certificate status request handler
func csGet(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	// Check Content-Type
	use_cbor := false
	switch r.Header.Get("Content-Type") {
	case "application/cbor":
		use_cbor = true
	case "application/json":
	case "":
		// Default to JSON if not Content-Type provided (curl, etc.)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request: Content-Type must be application/cbor or application/json"}`))
		return
	}

	// Parse serial number in request
	serialNumber, ok := pathParams["serial"]
	ser := new(big.Int)
	ser, ok = ser.SetString(serialNumber, 10)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid request"}`))
		return
	}

	// Check DB for serial number
	valid, err := db.SerialValid(ser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid serial number"}`))
		fmt.Println("cs:", err)
		return
	}

	// Convert bool to int for status field
	var valint int
	if valid {
		valint = 1
	} else {
		valint = 0
	}

	// Content-Type response handlers
	if use_cbor {
		// fmt.Printf("cs: got cert status request in CBOR for serial: %s\n", ser)
		w.Header().Set("Content-Type", "application/cbor")
		w.WriteHeader(http.StatusOK)
		enc := cbor.NewEncoder(w)
		err = enc.Encode(&protocol.CertStatusResponse{
			Status: valint,
		})
	} else {
		// fmt.Printf("cs: got cert status request in JSON for serial: %s\n", ser)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		err = enc.Encode(&protocol.CertStatusResponse{
			Status: valint,
		})
	}
}

// Test endpoint: https://localhost/api/v1/ds/{uuid}
func dsGet(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	// Check Content-Type request
	use_cbor := false
	switch r.Header.Get("Content-Type") {
	case "application/cbor":
		use_cbor = true
	case "application/json":
	case "":
		// Default to JSON if not Content-Type provided (curl, etc.)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request: Content-Type must be application/cbor or application/json"}`))
		return
	}

	// Parse UUID from request
	var err error
	var devid uuid.UUID
	if val, ok := pathParams["uuid"]; ok {
		devid, err = uuid.Parse(val)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "need a valid UUID"}`))
			return
		}
	}

	// Check UUID for valid certs
	serials, err := db.CertsByUUID(devid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "unable to query db for UUID"}`))
		log.Fatal("DB error: %s\n", err)
		return
	}

	// CBOR response handler
	if use_cbor {
		// fmt.Printf("Got dev status request in CBOR for UUID: %s\n", devid)
		w.Header().Set("Content-Type", "application/cbor")
		w.WriteHeader(http.StatusOK)
		enc := cbor.NewEncoder(w)
		if serials == nil {
			err = enc.Encode(&protocol.DevStatusResponse{
				Status:  0,
				Serials: nil,
			})
		} else {
			err = enc.Encode(&protocol.DevStatusResponse{
				Status:  1,
				Serials: serials,
			})
		}
		return
	}

	// Default JSON response handler
	// fmt.Printf("Got dev status request in JSON for UUID: %s\n", devid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	if serials == nil {
		err = enc.Encode(&protocol.DevStatusResponse{
			Status:  0,
			Serials: nil,
		})
	} else {
		err = enc.Encode(&protocol.DevStatusResponse{
			Status:  1,
			Serials: serials,
		})
	}
}

// Key update request handler
func kurPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "kur POST called"}`))

	// TODO: Validate current cert status and update/regen if necessary
	// This should generate a new certificate for this client,
	// based on the information from the existing certificate.
	// This should _not_ invalidate or revoke the old certificate,
	// as something such as an untimely power loss would cause the
	// new key to be lost.  The old certificate is fine to use
	// until it expires.
}

// Key revocation request
func krrPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "krr POST called"}`))

	// TODO: Mark certificate as revoked in the DB
}

// Test endpoint: https://localhost/api/v1/ccs
func ccsGet(w http.ResponseWriter, r *http.Request) {
	// Check Content-Type request
	use_cbor := false
	switch r.Header.Get("Content-Type") {
	case "application/cbor":
		use_cbor = true
	case "application/json":
	case "":
		// Default to JSON if not Content-Type provided (curl, etc.)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request: Content-Type must be application/cbor or application/json"}`))
		return
	}

	hubname := viper.GetString("server.hubname")
	port := viper.GetInt("server.mqttport")

	// CBOR response handler
	if use_cbor {
		// fmt.Printf("Got dev status request in CBOR for UUID: %s\n", devid)
		w.Header().Set("Content-Type", "application/cbor")
		w.WriteHeader(http.StatusOK)
		enc := cbor.NewEncoder(w)

		enc.Encode(&protocol.CCSResponse{
			Hubname: hubname,
			Port:    port,
		})
		return
	}

	// Default JSON response handler
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.Encode(&protocol.CCSResponse{
		Hubname: hubname,
		Port:    port,
	})
}

// REST API catch all handler
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "endpoint not found"}`))
}

// Root page handler
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Linaro CA HTTP server\n"))
}

// X509 Certificate copy handler
func ccGet(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	// Check Content-Type request
	use_cbor := false
	switch r.Header.Get("Content-Type") {
	case "application/cbor":
		use_cbor = true
	case "application/json":
	case "":
		// Default to JSON if not Content-Type provided (curl, etc.)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request: Content-Type must be application/cbor or application/json"}`))
		return
	}
	// Parse serial number in request
	serialNumber, ok := pathParams["serial"]
	ser := new(big.Int)
	ser, ok = ser.SetString(serialNumber, 10)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid request"}`))
		return
	}

	cert, err := db.GetCertBySerial(ser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid serial number"}`))
		fmt.Println("cs:", err)
		return
	}

	// Convert DER output to PEM
	pemout := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})

	// CBOR response handler
	if use_cbor {
		w.Header().Set("Content-Type", "application/cbor")
		w.WriteHeader(http.StatusOK)
		enc := cbor.NewEncoder(w)
		enc.Encode(protocol.CCResponse{
			Status: 0,
			Cert:   string(pemout),
		})
		fmt.Println(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(protocol.CCResponse{
		Status: 0,
		Cert:   string(pemout),
	})
}

// Start the HTTP Server
func Start(port int16) {
	var err error
	db, err = cadb.Open()
	if err != nil {
		log.Fatal("Unable to open CADB.db database")
	}

	// go registration()

	// Create a certificate pool with the CA certificate.
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("certs/CA.crt")
	if err != nil {
		log.Fatal(err)
	}
	certPool.AppendCertsFromPEM(caCert)

	r := mux.NewRouter()

	// Setup the REST API subrouter
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/cr", crPost).Methods(http.MethodPost)
	api.HandleFunc("/p10cr", p10crPost).Methods(http.MethodPost)
	api.HandleFunc("/cs/{serial}", csGet).Methods(http.MethodGet)
	api.HandleFunc("/ds/{uuid}", dsGet).Methods(http.MethodGet)
	api.HandleFunc("/kur", kurPost).Methods(http.MethodPost)
	api.HandleFunc("/krr", krrPost).Methods(http.MethodPost)
	api.HandleFunc("/ccs", ccsGet).Methods(http.MethodGet)
	api.HandleFunc("/cc/{serial}", ccGet).Methods(http.MethodGet)
	api.HandleFunc("", notFound)

	// Handle standard requests. Routes are tested in the order they are added,
	// so these will only be handled if they don't match anything above.
	r.HandleFunc("/", home)

	// Make sure the server key and certificate exist
	if !fileExists("certs/SERVER.key") || !fileExists("certs/SERVER.crt") {
		log.Fatal("Server certificate and key not found. See README.md.")
	}

	// Get the hostname from the $CAHOSTNAME environment variable
	hostname := os.Getenv("CAHOSTNAME")
	if hostname == "" {
		// Fall back to the system hostname (bash $HOSTNAME, zsh $HOST) if
		// nothing is defined.
		var err error
		hostname, err = os.Hostname()
		if (err != nil) || (hostname == "") {
			// As a last resort, fall back to localhost
			hostname = "localhost"
		}
	}

	server := &http.Server{
		// Use the system hostname.
		Addr:    hostname + ":" + strconv.Itoa(int(port)),
		Handler: r,

		// The certificate will be filled in by the listen and
		// server, but we can request/verify that there is a
		// valid client cert specified.
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,

			ClientCAs:             certPool,
			VerifyPeerCertificate: validatePeer,
		},
	}

	fmt.Println("Starting CA server on https://" + hostname + ":" +
		strconv.Itoa(int(port)))
	err = server.ListenAndServeTLS("certs/SERVER.crt", "certs/SERVER.key")
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

// ValidatePeer checks the given certificates and makes sure they are
// appropriate for requests from the bootstrap service.
func validatePeer(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(verifiedChains) != 1 {
		return fmt.Errorf("Expecting a single certificate chain")
	}

	// TODO: We should probably verify the certificate chain ends
	// with our CA, but that should always be the case.  In this
	// case, just verify the subject has an OU of "LinaroCA
	// Bootstrap Cert".
	//log.Printf("cert: %#v", verifiedChains[0][0].Subject)
	crt := verifiedChains[0][0]
	if len(crt.Subject.OrganizationalUnit) != 1 || crt.Subject.OrganizationalUnit[0] != "LinaroCA Bootstrap Cert" {
		return fmt.Errorf("Invalid client certificate")
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
