package httpserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/microbuilder/linaroca/protocol"
)

// Initialisation request handler
func irPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ir POST called"}
{"serial": "Device serial number"}`))
}

// Certification request handler
func crPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dec := json.NewDecoder(r.Body)
	var req protocol.CSRRequest
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
	err = enc.Encode(&protocol.CSRResponse{
		Status: 0,
		Cert:   cert,
	})
}

// Certification request from PKCS#10 handler
func p10crPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "p10cr POST called"}`))
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
}

// Key update request handler
func kurPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "kur POST called"}`))
}

// Key revocation request
func krrPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "krr POST called"}`))
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
	api.HandleFunc("/ir", irPost).Methods(http.MethodPost)
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

	fmt.Println("Starting HTTPS server on port https://localhost:" + strconv.Itoa(int(port)))
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
