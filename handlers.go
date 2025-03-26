package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var templates *template.Template

// LoadTemplates parses all HTML templates.
func LoadTemplates() {
	tmplPath := filepath.Join("web", "templates", "*.html")
	templates = template.Must(template.ParseGlob(tmplPath))
}

// indexHandler serves the main HTML page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch initial data to potentially render server-side (optional, JS will update)
	current, _ := GetCurrentStatus() // Ignore error for initial load, JS will handle missing data
	// history, _ := GetRecentStatusHistory() // Can pass this too if needed

	data := map[string]interface{}{
		"CurrentStatus": current,
		"TargetURL":     checkURL,
		"HistoryDays":   historyLimitDays,
		// Add uptime data if desired for initial render
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// apiStatusCurrentHandler serves the latest status as JSON.
func apiStatusCurrentHandler(w http.ResponseWriter, r *http.Request) {
	// Use the globally stored last status for immediate response
	if lastStatus == nil {
		// Fallback to DB if global state isn't populated yet
		var err error
		lastStatus, err = GetCurrentStatus()
		if err != nil {
			http.Error(w, "Error fetching current status", http.StatusInternalServerError)
			return
		}
		if lastStatus == nil {
			// Still no status
			http.Error(w, "No status data available yet", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lastStatus)
}

// apiStatusHistoryHandler serves recent historical data as JSON.
func apiStatusHistoryHandler(w http.ResponseWriter, r *http.Request) {
	history, err := GetRecentStatusHistory()
	if err != nil {
		http.Error(w, "Error fetching status history", http.StatusInternalServerError)
		return
	}

	if history == nil {
		history = []StatusRecord{} // Return empty array instead of null
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// apiUptimeHandler serves uptime percentages as JSON.
func apiUptimeHandler(w http.ResponseWriter, r *http.Request) {
	uptime24h, err24 := CalculateUptime(24 * time.Hour)
	uptime7d, err7 := CalculateUptime(7 * 24 * time.Hour)
	uptime30d, err30 := CalculateUptime(30 * 24 * time.Hour)

	// Basic error handling: log errors but still return best-effort data
	if err24 != nil {
		log.Printf("Error calculating 24h uptime: %v", err24)
		uptime24h = -1
	} // Use -1 to indicate error
	if err7 != nil {
		log.Printf("Error calculating 7d uptime: %v", err7)
		uptime7d = -1
	}
	if err30 != nil {
		log.Printf("Error calculating 30d uptime: %v", err30)
		uptime30d = -1
	}

	uptimeData := map[string]string{ // Format as strings for display
		"uptime24h": fmt.Sprintf("%.2f%%", uptime24h),
		"uptime7d":  fmt.Sprintf("%.2f%%", uptime7d),
		"uptime30d": fmt.Sprintf("%.2f%%", uptime30d),
	}
	if uptime24h < 0 {
		uptimeData["uptime24h"] = "Error"
	}
	if uptime7d < 0 {
		uptimeData["uptime7d"] = "Error"
	}
	if uptime30d < 0 {
		uptimeData["uptime30d"] = "Error"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uptimeData)
}
