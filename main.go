package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", statusHandler)
	fmt.Println("Server listening on port 8000")
	http.ListenAndServe(":8000", nil)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status, statusCode := checkStatus()
	data := struct {
		Status     string
		Color      string
		Timestamp  string
		StatusCode int
		StartTime  string
	}{
		Status:     status,
		Color:      getColor(status),
		Timestamp:  formatTime(time.Now().In(time.FixedZone("EAT", 3*60*60))),
		StatusCode: statusCode,
		StartTime:  formatTime(startTime),
	}

	tmpl, err := template.New("status").Parse(statusTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM MST")
}

var startTime = time.Now().In(time.FixedZone("EAT", 3*60*60))

func checkStatus() (string, int) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://deploy.tz")

	if err != nil {
		return "DOWN", 0
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "UP", resp.StatusCode
	} else if resp.StatusCode == 404 {
		return "Page Not Found", resp.StatusCode
	} else if resp.StatusCode == 500 {
		return "Internal Server Error", resp.StatusCode
	} else {
		return "DOWN", resp.StatusCode
	}

}

func getColor(status string) string {
	switch status {
	case "UP":
		return "green"
	case "Page Not Found":
		return "orange"
	case "Internal Server Error":
		return "red"
	default:
		return "red"
	}
}

const statusTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Deploy Status Monitor</title>
    <style>
        body {
            font-family: sans-serif;
            background-color: #f0f0f0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
        }

        .container {
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            text-align: center;
        }

        .status {
            font-size: 2em;
            font-weight: bold;
            color: {{.Color}}; /* Use color variable */
        }

        .green { color: green; }
        .orange { color: orange; }
        .red { color: red; }

        .timestamp, .statusCode, .startTime {
            color: #777;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
    <h1><a href="https://deploy.tz">Deploy</a> Status Monitor</h1>
        <p class="status">{{.Status}}</p>
        <p class="statusCode">HTTP Status Code: {{.StatusCode}}</p>
        <p class="timestamp">Last checked: {{.Timestamp}}</p>
        <p class="startTime">Monitoring active since {{.StartTime}}</p>
    </div>
</body>
</html>`
