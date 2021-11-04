package caserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

type CSRRequest struct {
	CSR []byte
}

type CSRResponse struct {
	Status int
	Cert   []byte
}

// Certification request handler
func crPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dec := json.NewDecoder(r.Body)
	var req CSRRequest
	err := dec.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request"}`))
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(&CSRResponse{
		Status: 0,
		Cert:   cert,
	})
}

// Maximum file size for uploaded CSRs = 4 KB
const MAX_CSR_UPLOAD_SIZE = 1024 * 4

// Certification request from PKCS#10 handler
func p10crPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

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
	fmt.Printf("CSR file size: %v bytes\n", fileSize)
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

	// TOOD: Parse input CSR file and reply with certificate in PEM format
	// MIME type = application/x-x509-user-cert or application/x-pem-file ?

	fmt.Printf("Got csr: \n%s\n", fileBytes)

	// cert, err := handleCSR(fileBytes)
	// if err != nil {
	// 	// TODO: Encode the error.
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte(`{"error": "Invalid CSR"}`))
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/x-pem-file")
	// w.WriteHeader(http.StatusOK)
	// w.Write(cert)
	// enc := json.NewEncoder(w)
	// err = enc.Encode(&CSRResponse{
	// 	Status: 0,
	// 	Cert:   cert,
	// })

	w.Write([]byte("SUCCESS"))
}

// Certificate status request handler
func csGet(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	if serialNumber, ok := pathParams["serial"]; ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"message": "cs GET called"}
{"serial": %s}`, serialNumber)))
		return
	}

	// TODO: Check DB for existence and state of specific serial
}

// Key update request handler
func kurPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "kur POST called"}`))

	// TODO: Validate current cert status and update/regen if necessary
}

// Key revocation request
func krrPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "krr POST called"}`))

	// TODO: Mark certificate as revoked in the DB
}

// RESET API catch all handler
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "endpoint not found"}`))
}

// Test endpoint: https://localhost/api/v1/status/123?format=cbor
func statusGet(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	deviceID := -1
	var err error
	if val, ok := pathParams["deviceID"]; ok {
		deviceID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	query := r.URL.Query()
	format := query.Get("format")
	if len(format) > 0 {
		w.Write([]byte(fmt.Sprintf(`{"deviceID": %d, "format":, "%s"}`, deviceID, format)))
	} else {
		w.Write([]byte(fmt.Sprintf(`{"deviceID": %d}`, deviceID)))
	}
}

// Root page handler
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Linaro CA HTTP server\n"))
}

// Start the HTTP Server
func Start(port int16) {
	r := mux.NewRouter()

	// Setup the REST API subrouter
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/cr", crPost).Methods(http.MethodPost)
	api.HandleFunc("/p10cr", p10crPost).Methods(http.MethodPost)
	api.HandleFunc("/cs/{serial}", csGet).Methods(http.MethodGet)
	api.HandleFunc("/kur", kurPost).Methods(http.MethodPost)
	api.HandleFunc("/krr", krrPost).Methods(http.MethodPost)
	api.HandleFunc("/status/{deviceID}", statusGet).Methods(http.MethodGet)
	api.HandleFunc("", notFound)

	// Handle standard requests. Routes are tested in the order they are added,
	// so these will only be handled if they don't match anything above.
	r.HandleFunc("/", home)

	// Make sure the server key and certificate exist
	if !fileExists("SERVER.key") || !fileExists("SERVER.crt") {
		log.Fatal("Server certificate and key not found. See README.md.")
	}

	fmt.Println("Starting CA server on port https://localhost:" + strconv.Itoa(int(port)))
	err := http.ListenAndServeTLS(":"+strconv.Itoa(int(port)), "SERVER.crt", "SERVER.key", r)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
