package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var port int
	var dir string

	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&dir, "dir", "./public", "Directory to serve")
	flag.Parse()

	// Validate directory
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", dir)
	}

	http.Handle("/", http.FileServer(http.Dir(dir)))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("GhostGate serving %s on http://localhost%s\n", dir, addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
