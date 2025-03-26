package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Setup logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize Database
	InitDB()
	defer db.Close()

	// Load HTML Templates
	LoadTemplates()

	// Start Background Checker
	StartChecker()

	// Setup HTTP Routes
	mux := http.NewServeMux()

	// Serve Static Files (CSS, JS)
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Page Handler
	mux.HandleFunc("/", indexHandler)

	// API Handlers
	mux.HandleFunc("/api/status/current", apiStatusCurrentHandler)
	mux.HandleFunc("/api/status/history", apiStatusHistoryHandler)
	mux.HandleFunc("/api/status/uptime", apiUptimeHandler)

	// Start Server
	port := "0.0.0.0:8000" // Make configurable if needed (e.g., via env var)
	log.Printf("Starting status page server on %s", port)

	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", port, err)
	}
}
