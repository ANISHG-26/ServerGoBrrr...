package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Status values used to simulate normal operation, soft degradation, and a full failure spiral.
const (
	StatusOK          = "ok"
	StatusDegraded    = "degraded"
	StatusMadaoSpiral = "madao_spiral"
)

// HealthResponse is returned by GET /health so consumers can reason about the current state.
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	LatencyMS int64  `json:"latency_ms"`
	Time      string `json:"time"`

	MessageJP string `json:"message_jp"`
	MessageEN string `json:"message_en"`
	RequestID string `json:"request_id"`
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// madaoLine maps each health state to a thematic line in Japanese and English.
func madaoLine(status string) (jp string, en string) {
	switch status {
	case StatusOK:
		return "まだ終わっちゃいねぇ…！ (マダオ)", "Not done yet. Still standing."
	case StatusDegraded:
		return "心が折れそうだ…でも動く (マダオ)", "Slow and struggling… but still responding."
	case StatusMadaoSpiral:
		return "マダオスパイラル突入…それでも立ち上がる！ (マダオ)", "Entered the Madao Spiral… but I'll get back up."
	default:
		return "なんとかなるさ…たぶん (マダオ)", "It’ll work out… probably."
	}
}

func makeRequestID() string {
	return fmt.Sprintf("madao-%06d", rand.Intn(1_000_000))
}

func main() {
	// Service identity
	port := flag.Int("port", 8080, "Port to run the server on")
	name := flag.String("name", "Madao API", "Name of the service")
	version := flag.String("version", "0.1.0", "Service version")

	// Chaos controls
	chaos := flag.Bool("chaos", false, "Enable chaos mode (latency + failures)")
	failPct := flag.Int("fail-pct", 20, "Chance to fail health check (0-100) in chaos mode")
	slowPct := flag.Int("slow-pct", 30, "Chance the request becomes slow (0-100) in chaos mode")
	maxDelayMS := flag.Int("max-delay-ms", 1200, "Max added delay in ms when slowness triggers")
	slowThresholdMS := flag.Int("slow-threshold-ms", 800, "If latency >= this, mark as degraded (unless failing)")
	persistent := flag.Bool("persistent", true, "If true, failures still return a diagnostic JSON body (useful for drills)")

	// Graceful shutdown knob: how long we wait for in-flight requests to finish
	shutdownTimeoutSec := flag.Int("shutdown-timeout-sec", 5, "Seconds to wait for graceful shutdown")

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := makeRequestID()

		status := StatusOK
		httpCode := http.StatusOK
		slowTriggered := false

		if *chaos {
			if rand.Intn(100) < *slowPct {
				slowTriggered = true
				delay := time.Duration(rand.Intn(*maxDelayMS+1)) * time.Millisecond
				time.Sleep(delay)
			}

			if rand.Intn(100) < *failPct {
				status = StatusMadaoSpiral
				httpCode = http.StatusServiceUnavailable

				// When persistence is off, return a minimal payload to mimic outages with scarce diagnostics.
				if !*persistent {
					writeJSON(w, httpCode, map[string]any{
						"status":     status,
						"service":    *name,
						"version":    *version,
						"request_id": reqID,
					})
					log.Printf("[%s] %s %s -> %s (minimal) code=%d", reqID, r.Method, r.URL.Path, status, httpCode)
					return
				}
			}
		}

		latency := time.Since(start).Milliseconds()

		// Treat slow-but-responding requests as degraded so clients can exercise fallback logic.
		if status != StatusMadaoSpiral && latency >= int64(*slowThresholdMS) {
			status = StatusDegraded
			httpCode = http.StatusOK
		}

		jp, en := madaoLine(status)

		resp := HealthResponse{
			Status:    status,
			Service:   *name,
			Version:   *version,
			LatencyMS: latency,
			Time:      time.Now().Format(time.RFC3339),
			MessageJP: jp,
			MessageEN: en,
			RequestID: reqID,
		}

		writeJSON(w, httpCode, resp)
		log.Printf("[%s] %s %s -> %s (%dms) code=%d slowTriggered=%v",
			reqID, r.Method, r.URL.Path, status, latency, httpCode, slowTriggered)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Welcome to %s (version: %s)\n", *name, *version)
		fmt.Fprintln(w, "Try: GET /health")
	})

	addr := fmt.Sprintf(":%d", *port)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown: start server in background
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting %s v%s on http://localhost%s", *name, *version, addr)
		log.Printf("Chaos=%v failPct=%d slowPct=%d maxDelayMS=%d slowThresholdMS=%d persistent=%v",
			*chaos, *failPct, *slowPct, *maxDelayMS, *slowThresholdMS, *persistent)

		// ListenAndServe returns when server stops.
		// http.ErrServerClosed is expected during shutdown.
		if err := server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	// Graceful shutdown: listen for Ctrl+C / termination
	sigCh := make(chan os.Signal, 1)

	// On Windows: os.Interrupt (Ctrl+C) works.
	// On Linux/macOS: SIGTERM is common in containers/Kubernetes.
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("受信: %v — マダオ、静かに終了します… (Shutdown starting)", sig)

		// Give in-flight requests a deadline to finish
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*shutdownTimeoutSec)*time.Second)
		defer cancel()

		// Stop accepting new connections, wait for in-flight handlers to return (until timeout)
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("シャットダウン失敗: %v (forced exit)", err)
		} else {
			log.Printf("シャットダウン完了。おつかれさま。 (Shutdown complete)")
		}

	case err := <-errCh:
		// If server died unexpectedly, surface it.
		if err == http.ErrServerClosed {
			log.Printf("Server closed (expected).")
		} else {
			log.Fatalf("Server error: %v", err)
		}
	}

	// Optional: ensure the server is fully closed even after Shutdown.
	_ = server.Close()
}
