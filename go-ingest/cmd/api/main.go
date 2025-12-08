package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ColonelToad/EdgeSight/go-ingest/internal/embeddings"
	"github.com/ColonelToad/EdgeSight/go-ingest/internal/store"
)

func main() {
	// Initialize database
	dbPath := os.Getenv("EDGESIGHT_DB_PATH")
	if dbPath == "" {
		dbPath = "edgesight.db"
	}

	db, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Embedding sidecar client (optional)
	embedEndpoint := os.Getenv("EMBEDDING_ENDPOINT")
	if embedEndpoint == "" {
		embedEndpoint = "http://localhost:9000"
	}
	var embedCli *embeddings.Client
	if embedEndpoint != "" {
		embedCli = embeddings.NewClient(embedEndpoint)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("EdgeSight API Server starting on port %s", port)
	apiServer := NewAPIServer(db, embedCli)
	log.Fatal(http.ListenAndServe(":"+port, apiServer.Router()))
}

// APIServer holds the database connection and HTTP handlers
type APIServer struct {
	store       *store.SQLiteStore
	embedClient *embeddings.Client
}

// NewAPIServer creates a new API server instance
func NewAPIServer(db *store.SQLiteStore, embedCli *embeddings.Client) *APIServer {
	return &APIServer{store: db, embedClient: embedCli}
}

// Router configures all HTTP routes
func (s *APIServer) Router() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// Snapshot endpoints
	mux.HandleFunc("/api/v1/snapshots/latest", s.handleGetLatestSnapshot)
	mux.HandleFunc("/api/v1/snapshots/range", s.handleGetSnapshotsByRange)
	mux.HandleFunc("/api/v1/snapshots", s.handleGetSnapshots)

	// Metrics endpoints
	mux.HandleFunc("/api/v1/metrics/series", s.handleGetMetricSeries)

	// Embedding search / query
	mux.HandleFunc("/api/v1/search", s.handleSearch)
	mux.HandleFunc("/api/v1/query", s.handleQuery)

	// CORS and logging middleware
	return enableCORS(loggingMiddleware(mux))
}

// handleHealth returns API health status
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	respondJSON(w, http.StatusOK, response)
}

// handleSearch returns top similar snapshot summaries for a query.
func (s *APIServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	location := r.URL.Query().Get("location")
	if location == "" {
		location = "Los Angeles"
	}
	if q == "" {
		http.Error(w, "missing q", http.StatusBadRequest)
		return
	}
	if s.embedClient == nil {
		http.Error(w, "embedding service not configured", http.StatusServiceUnavailable)
		return
	}
	vec, err := s.embedClient.Embed(q)
	if err != nil {
		http.Error(w, fmt.Sprintf("embed error: %v", err), http.StatusBadGateway)
		return
	}
	results, err := s.store.SearchEmbeddings(location, vec, 5)
	if err != nil {
		http.Error(w, fmt.Sprintf("search error: %v", err), http.StatusInternalServerError)
		return
	}

	type res struct {
		Summary    string  `json:"summary"`
		SnapshotTS string  `json:"snapshot_ts"`
		Location   string  `json:"location"`
		Score      float64 `json:"score"`
	}
	out := make([]res, 0, len(results))
	for _, r := range results {
		out = append(out, res{
			Summary:    r.Summary,
			SnapshotTS: r.SnapshotTS,
			Location:   r.Location,
			Score:      r.Score,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": out,
	})
}

// handleQuery performs search then (placeholder) LLM answer.
func (s *APIServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	location := r.URL.Query().Get("location")
	if location == "" {
		location = "Los Angeles"
	}
	if q == "" {
		http.Error(w, "missing q", http.StatusBadRequest)
		return
	}
	if s.embedClient == nil {
		http.Error(w, "embedding service not configured", http.StatusServiceUnavailable)
		return
	}
	vec, err := s.embedClient.Embed(q)
	if err != nil {
		http.Error(w, fmt.Sprintf("embed error: %v", err), http.StatusBadGateway)
		return
	}
	results, err := s.store.SearchEmbeddings(location, vec, 5)
	if err != nil {
		http.Error(w, fmt.Sprintf("search error: %v", err), http.StatusInternalServerError)
		return
	}

	type src struct {
		Summary    string  `json:"summary"`
		SnapshotTS string  `json:"snapshot_ts"`
		Location   string  `json:"location"`
		Score      float64 `json:"score"`
	}
	sources := make([]src, 0, len(results))
	for _, r := range results {
		sources = append(sources, src{
			Summary:    r.Summary,
			SnapshotTS: r.SnapshotTS,
			Location:   r.Location,
			Score:      r.Score,
		})
	}

	answer := "LLM not configured; showing similar snapshots."
	if s.embedClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
		defer cancel()

		var sb strings.Builder
		sb.WriteString("Question: ")
		sb.WriteString(q)
		sb.WriteString("\nLocation: ")
		sb.WriteString(location)
		sb.WriteString("\nTop snapshots:\n")
		for i, src := range sources {
			sb.WriteString(fmt.Sprintf("%d) [%s] %s (score %.3f)\n", i+1, src.SnapshotTS, src.Summary, src.Score))
		}
		sb.WriteString("Provide a concise answer (<=3 sentences). If the context is insufficient, say so briefly.")

		systemPrompt := "You are EdgeSight's analyst. You summarize local conditions using provided snapshots only. Be concise, avoid speculation, and mention timestamps/metrics when relevant. If data is insufficient, say so."

		// Call Python sidecar /query endpoint
		queryPayload := map[string]interface{}{
			"system":     systemPrompt,
			"user":       sb.String(),
			"max_tokens": 256,
		}
		body, _ := json.Marshal(queryPayload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:9000/query", bytes.NewReader(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 45 * time.Second}
			if resp, err := client.Do(req); err == nil && resp.StatusCode == http.StatusOK {
				var result struct {
					Answer string `json:"answer"`
				}
				json.NewDecoder(resp.Body).Decode(&result)
				resp.Body.Close()
				answer = strings.TrimSpace(result.Answer)
			} else {
				if err != nil {
					answer = fmt.Sprintf("LLM error: %v", err)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"answer":  answer,
		"sources": sources,
	})
}

// handleGetLatestSnapshot returns the most recent snapshot for a location
func (s *APIServer) handleGetLatestSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	location := r.URL.Query().Get("location")
	if location == "" {
		location = "Los Angeles" // Default location
	}

	snapshot, err := s.store.GetLatestSnapshot(location)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch snapshot: "+err.Error())
		return
	}

	if snapshot == nil {
		respondError(w, http.StatusNotFound, "No snapshot found for location: "+location)
		return
	}

	respondJSON(w, http.StatusOK, snapshot)
}

// handleGetSnapshotsByRange returns snapshots within a time range
func (s *APIServer) handleGetSnapshotsByRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	location := r.URL.Query().Get("location")
	if location == "" {
		location = "Los Angeles"
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		respondError(w, http.StatusBadRequest, "Missing required parameters: start and end")
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid start time format. Use RFC3339: "+err.Error())
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid end time format. Use RFC3339: "+err.Error())
		return
	}

	snapshots, err := s.store.GetSnapshotsByTimeRange(location, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch snapshots: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"location": location,
		"start":    start.Format(time.RFC3339),
		"end":      end.Format(time.RFC3339),
		"count":    len(snapshots),
		"data":     snapshots,
	}

	respondJSON(w, http.StatusOK, response)
}

// handleGetSnapshots returns recent snapshots with pagination
func (s *APIServer) handleGetSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	location := r.URL.Query().Get("location")
	if location == "" {
		location = "Los Angeles"
	}

	// Default to last 24 hours
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 {
			hours = parsed
		}
	}

	end := time.Now().UTC()
	start := end.Add(-time.Duration(hours) * time.Hour)

	snapshots, err := s.store.GetSnapshotsByTimeRange(location, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch snapshots: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"location": location,
		"hours":    hours,
		"count":    len(snapshots),
		"data":     snapshots,
	}

	respondJSON(w, http.StatusOK, response)
}

// handleGetMetricSeries returns time series data for a specific metric
func (s *APIServer) handleGetMetricSeries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metric := r.URL.Query().Get("metric")
	location := r.URL.Query().Get("location")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if metric == "" {
		respondError(w, http.StatusBadRequest, "Missing required parameter: metric")
		return
	}

	if location == "" {
		location = "Los Angeles"
	}

	// Default to last 7 days if not specified
	var start, end time.Time
	if startStr == "" || endStr == "" {
		end = time.Now().UTC()
		start = end.Add(-7 * 24 * time.Hour)
	} else {
		var err error
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid start time format: "+err.Error())
			return
		}

		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end time format: "+err.Error())
			return
		}
	}

	series, err := s.store.GetMetricSeries(metric, location, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch metric series: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"metric":   metric,
		"location": location,
		"start":    start.Format(time.RFC3339),
		"end":      end.Format(time.RFC3339),
		"count":    len(series),
		"data":     series,
	}

	respondJSON(w, http.StatusOK, response)
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// enableCORS adds CORS headers to allow frontend access
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
