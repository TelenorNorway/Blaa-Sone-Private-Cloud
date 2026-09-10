package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Checker struct {
	client      *http.Client
	frontendURL string
}

type StatusResponse struct {
	API     string `json:"api"`
	Network string `json:"network"`
	Message string `json:"message"`
}

func (c *Checker) checkFrontend() error {
	response, err := c.client.Get(c.frontendURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("frontend returned HTTP %d", response.StatusCode)
	}

	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

func (c *Checker) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := c.checkFrontend(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (c *Checker) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := c.checkFrontend(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)

		json.NewEncoder(w).Encode(StatusResponse{
			API:     "ok",
			Network: "error",
			Message: "Backend cannot reach the frontend service.",
		})
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(StatusResponse{
		API:     "ok",
		Network: "ok",
		Message: "Backend can reach the frontend service. ",
	})
}

func main() {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://frontend-service"
	}

	checker := &Checker{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		frontendURL: frontendURL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", checker.readinessHandler)
	mux.HandleFunc("GET /api/status", checker.statusHandler)

	server := &http.Server{
		Addr:              ":8000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Backend listening on http://localhost:8000")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
