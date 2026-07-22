package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type config struct {
	Name         string
	Version      string
	Message      string
	Secret       string
	PodName      string
	PodNamespace string
	PodIP        string
	Port         string
	ReadyAfter   time.Duration
}

type response struct {
	App              string `json:"app"`
	Version          string `json:"version"`
	Message          string `json:"message"`
	PodName          string `json:"podName,omitempty"`
	PodNamespace     string `json:"podNamespace,omitempty"`
	PodIP            string `json:"podIP,omitempty"`
	SecretConfigured bool   `json:"secretConfigured"`
	Time             string `json:"time"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := loadConfig()
	readyAt := time.Now().Add(cfg.ReadyAfter)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, response{
			App:              cfg.Name,
			Version:          cfg.Version,
			Message:          cfg.Message,
			PodName:          cfg.PodName,
			PodNamespace:     cfg.PodNamespace,
			PodIP:            cfg.PodIP,
			SecretConfigured: cfg.Secret != "",
			Time:             time.Now().UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if time.Now().Before(readyAt) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "warming-up"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	server := &http.Server{Addr: ":" + cfg.Port, Handler: loggingMiddleware(logger, mux)}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server started", "addr", server.Addr, "ready_after", cfg.ReadyAfter)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("server shutting down")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func loadConfig() config {
	readyAfter := 0 * time.Second
	if raw := os.Getenv("READY_AFTER_SECONDS"); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err == nil && seconds >= 0 {
			readyAfter = time.Duration(seconds) * time.Second
		}
	}
	return config{
		Name:         valueOr("APP_NAME", "playground-api"),
		Version:      valueOr("APP_VERSION", "v1"),
		Message:      valueOr("APP_MESSAGE", "hello from Kubernetes"),
		Secret:       os.Getenv("SECRET_MESSAGE"),
		PodName:      os.Getenv("POD_NAME"),
		PodNamespace: os.Getenv("POD_NAMESPACE"),
		PodIP:        os.Getenv("POD_IP"),
		Port:         valueOr("PORT", "8080"),
		ReadyAfter:   readyAfter,
	}
}

func valueOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started).String())
	})
}
