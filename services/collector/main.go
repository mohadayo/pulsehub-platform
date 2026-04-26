package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Metric struct {
	Service   string    `json:"service"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type MetricStore struct {
	mu      sync.RWMutex
	metrics []Metric
}

func NewMetricStore() *MetricStore {
	return &MetricStore{metrics: make([]Metric, 0)}
}

func (s *MetricStore) Add(m Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m.Timestamp = time.Now().UTC()
	s.metrics = append(s.metrics, m)
}

func (s *MetricStore) List() []Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Metric, len(s.metrics))
	copy(result, s.metrics)
	return result
}

type Server struct {
	store  *MetricStore
	logger *log.Logger
}

func NewServer(store *MetricStore, logger *log.Logger) *Server {
	return &Server{store: store, logger: logger}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "collector"})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.logger.Println("GET /metrics")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.store.List())
	case http.MethodPost:
		var m Metric
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			s.logger.Printf("ERROR: invalid metric payload: %v", err)
			http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
			return
		}
		if m.Service == "" || m.Name == "" {
			s.logger.Println("ERROR: missing required fields service or name")
			http.Error(w, `{"error":"service and name are required"}`, http.StatusBadRequest)
			return
		}
		s.store.Add(m)
		s.logger.Printf("POST /metrics service=%s name=%s value=%.2f", m.Service, m.Name, m.Value)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func SetupRoutes(srv *Server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/metrics", srv.handleMetrics)
	return mux
}

func main() {
	port := os.Getenv("COLLECTOR_PORT")
	if port == "" {
		port = "8081"
	}
	logger := log.New(os.Stdout, "[collector] ", log.LstdFlags)
	store := NewMetricStore()
	srv := NewServer(store, logger)
	mux := SetupRoutes(srv)

	logger.Printf("Starting collector service on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
