package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Configure the logger
	log.SetFlags(log.Lshortfile | log.Ltime)

	// Check the input arguments
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <path-to-config>\n", os.Args[0])
	}

	// Load the configuration
	conf, err := loadConf(os.Args[1])
	if err != nil {
		log.Fatalf("error loading the configuration: %v\n", err)
	}

	// Instantiate a new server
	s := newServer()

	// Initialize handlers
	http.HandleFunc("/", cors(s.root))
	http.HandleFunc("/search", cors(s.search))
	http.HandleFunc("/query", cors(s.query))
	http.HandleFunc("/annotations", cors(s.annotations))

	// Start the server
	log.Printf("starting the server...\n")
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", conf.BindAddr, conf.BindPort), nil); err != nil {
		log.Fatalf("couldn't bind the server: %v\n", err)
	}
}
