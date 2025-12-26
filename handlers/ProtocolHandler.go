package Handlers

import (
	"fmt"
	"net/http"
)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HTTP Traffic - Received on Port 80 or Proxy\n")
}

// HTTPS Handler (TLS handler)
func httpsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HTTPS Traffic - Received on Port 443 or Proxy\n")
}
