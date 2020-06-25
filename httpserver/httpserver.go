package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get called"}`))
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "post called"}`))
}

func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message": "put called"}`))
}

func delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "delete called"}`))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}

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

// Start the HTTP Server
func Start(port int16) {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("", get).Methods(http.MethodGet)
	api.HandleFunc("", post).Methods(http.MethodGet)
	api.HandleFunc("", put).Methods(http.MethodGet)
	api.HandleFunc("", delete).Methods(http.MethodGet)
	api.HandleFunc("", notFound)

	// Can be tested with a GET call to https://localhost/api/v1/status/123?format=cbor
	api.HandleFunc("/status/{deviceID}", statusGet).Methods(http.MethodGet)

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

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Linaro CA HTTP server.\n"))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
