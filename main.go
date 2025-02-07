package main

import (
	"html/template"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", statusHandler)
	http.ListenAndServe(":8000", nil)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status, statusCode := checkStatus()
	data := struct {
		Status     string
		Color      string
		Timestamp  string
		StatusCode int
	}{
		Status:     status,
		Color:      getColor(status),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		StatusCode: statusCode,
	}

	tmpl := template.Must(template.New("status").Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>deploy.tz Status</title>
		<style>
			body { font-family: Arial, sans-serif; text-align: center; padding: 2rem; }
			.status { font-size: 1.5rem; margin: 1rem 0; }
			.up { color: #2ecc71; }
			.down { color: #e74c3c; }
			.footer { color: #7f8c8d; margin-top: 2rem; }
		</style>
	</head>
	<body>
		<h1>deploy.tz Status Monitor</h1>
		<div class="status">
			Status: <span class="{{.Color}}">{{.Status}}</span>
		</div>
		<div>HTTP Status Code: {{.StatusCode}}</div>
		<div class="footer">
			Last checked: {{.Timestamp}}<br>
			Monitoring active since {{.StartTime}}
		</div>
	</body>
	</html>
	`))
	
	data.StartTime = startTime.Format(time.RFC3339)
	tmpl.Execute(w, data)
}

var startTime = time.Now()

func checkStatus() (string, int) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://deploy.tz")
	
	if err != nil {
		return "DOWN", 0
	}
	defer resp.Body.Close()
	
	return "UP", resp.StatusCode
}

func getColor(status string) string {
	if status == "UP" {
		return "up"
	}
	return "down"
}
