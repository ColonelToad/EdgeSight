package models

import "time"

// CanonicalData represents the unified data structure for all ingested data
type CanonicalData struct {
	ID        int64     `json:"id"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}
