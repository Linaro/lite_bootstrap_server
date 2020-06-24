package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func Start() {
	// Make sure the server key and certificate exist
	if !fileExists("SERVER.key") || !fileExists("SERVER.crt") {
		log.Fatal("Server certificate and key not found. See README.md.")
	}
	fmt.Println("Starting HTTPS server on port https://localhost:443")
	http.HandleFunc("/", HelloServer)
	err := http.ListenAndServeTLS(":443", "SERVER.crt", "SERVER.key", nil)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("CA HTTP server.\n"))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
