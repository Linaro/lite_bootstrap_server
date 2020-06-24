package httpserver

import (
	"fmt"
	"net/http"
)

func Start() {
	http.HandleFunc("/", HelloServer)
	fmt.Println("Starting HTTPS server on port 8080")
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "success!")
}
