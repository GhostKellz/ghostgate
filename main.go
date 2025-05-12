package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	var port int
	var dir string
	var backend string

	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&dir, "dir", "./public", "Directory to serve")
	flag.StringVar(&backend, "backend", "", "Backend server URL for reverse proxy")
	flag.Parse()

	// Validate directory
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", dir)
	}

	// Set up handlers
	if backend != "" {
		backendURL, err := url.Parse(backend)
		if err != nil {
			log.Fatalf("Invalid backend URL: %v", err)
		}

		proxy := httputil.NewSingleHostReverseProxy(backendURL)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
		log.Printf("Reverse proxying to backend: %s", backend)
	} else {
		http.Handle("/", http.FileServer(http.Dir(dir)))
		log.Printf("Serving static files from: %s", dir)
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("GhostGate running on http://localhost%s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
