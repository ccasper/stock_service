package common

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	busyness float64
	version  string
	uptime   time.Time
)

func StartHealthServer(newVersion string, port string) error {
	log.Printf("Health port starting on http:/%s/health", port)

	uptime = time.Now()
	version = newVersion

	if !strings.Contains(port, ":") {
		return fmt.Errorf("invalid port: %s, needs a ':<port>'", port)
	}

	// Create a new ServeMux for this server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	// Run the server in a goroutine so it doesn't block main
	go func() {
		fmt.Printf("Starting server on %s\n", port)
		if err := http.ListenAndServe(port, mux); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()

	return nil
}

func SetBusyness(newBusyness float64) {
	busyness = newBusyness
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "busyness=%.1f\nversion=%s\nuptime=%.1f\n", busyness, version, time.Since(uptime).Seconds())
}
