package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReadyEndpointWarmsUpThenBecomesReady(t *testing.T) {
	readyAt := time.Now().Add(20 * time.Millisecond)
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if time.Now().Before(readyAt) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "warming-up"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if first.Code != http.StatusServiceUnavailable {
		t.Fatalf("first readiness status = %d, want %d", first.Code, http.StatusServiceUnavailable)
	}

	time.Sleep(25 * time.Millisecond)
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if second.Code != http.StatusOK {
		t.Fatalf("second readiness status = %d, want %d", second.Code, http.StatusOK)
	}
}
