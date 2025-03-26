package main

import (
	"log"
	"net/http"
	"time"
)

const (
	checkURL      = "https://deploy.tz"
	checkInterval = 1 * time.Minute
	checkTimeout  = 10 * time.Second // Timeout for the HTTP request
)

var (
	// Store the last known status globally for quick access
	lastStatus *StatusRecord
)

// StartChecker runs the status check periodically in a background goroutine.
func StartChecker() {
	log.Printf("Starting status checker for %s every %v", checkURL, checkInterval)
	ticker := time.NewTicker(checkInterval)

	// Run once immediately
	go checkStatus()

	// Then run on schedule
	go func() {
		for range ticker.C {
			checkStatus()
		}
	}()
}

// checkStatus performs a single status check for the target URL.
func checkStatus() {
	client := &http.Client{
		Timeout: checkTimeout,
	}

	start := time.Now()
	isUp := false
	statusCode := 0 // Default to 0 if request fails completely

	resp, err := client.Get(checkURL)
	responseTimeMs := time.Since(start).Milliseconds()

	if err != nil {
		log.Printf("Check failed for %s: %v", checkURL, err)
		// isUp remains false
	} else {
		defer resp.Body.Close()
		statusCode = resp.StatusCode
		// Consider any 2xx or 3xx status as "UP"
		if statusCode >= 200 && statusCode < 400 {
			isUp = true
			log.Printf("Check successful for %s: Status %d, Response Time %dms", checkURL, statusCode, responseTimeMs)
		} else {
			log.Printf("Check failed for %s: Status %d", checkURL, statusCode)
			// isUp remains false
		}
	}

	// Update global last status
	newStatus := &StatusRecord{
		Timestamp:      time.Now(),
		IsUp:           isUp,
		ResponseTimeMs: responseTimeMs,
		StatusCode:     statusCode,
	}
	lastStatus = newStatus // Update global state

	// Save to database
	err = AddStatusRecord(isUp, responseTimeMs, statusCode)
	if err != nil {
		log.Printf("Failed to save status record: %v", err)
	}
}
