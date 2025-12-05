package models

import "time"

// RawData represents the raw data structure from API responses
type RawData struct {
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
