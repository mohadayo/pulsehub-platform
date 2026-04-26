package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestServer() *Server {
	store := NewMetricStore()
	logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
	return NewServer(store, logger)
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()
	mux := SetupRoutes(srv)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", resp["status"])
	}
	if resp["service"] != "collector" {
		t.Fatalf("expected service collector, got %s", resp["service"])
	}
}

func TestPostMetric(t *testing.T) {
	srv := newTestServer()
	mux := SetupRoutes(srv)

	body := `{"service":"web","name":"cpu_usage","value":42.5}`
	req := httptest.NewRequest(http.MethodPost, "/metrics", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestGetMetrics(t *testing.T) {
	srv := newTestServer()
	mux := SetupRoutes(srv)

	body := `{"service":"web","name":"memory","value":1024}`
	req := httptest.NewRequest(http.MethodPost, "/metrics", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var metrics []Metric
	json.NewDecoder(w.Body).Decode(&metrics)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Service != "web" {
		t.Fatalf("expected service web, got %s", metrics[0].Service)
	}
}

func TestPostMetricInvalidPayload(t *testing.T) {
	srv := newTestServer()
	mux := SetupRoutes(srv)

	req := httptest.NewRequest(http.MethodPost, "/metrics", bytes.NewBufferString(`{invalid`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPostMetricMissingFields(t *testing.T) {
	srv := newTestServer()
	mux := SetupRoutes(srv)

	req := httptest.NewRequest(http.MethodPost, "/metrics", bytes.NewBufferString(`{"value":10}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := newTestServer()
	mux := SetupRoutes(srv)

	req := httptest.NewRequest(http.MethodDelete, "/metrics", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
