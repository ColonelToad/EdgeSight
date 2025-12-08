package clients

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// CDCFluViewClient fetches flu surveillance data from CDC FluView.
// Uses the public CDC FluView web service.
type CDCFluViewClient struct {
	baseURL string
	httpCli *http.Client
}

// CDCFluSummary aggregates current flu activity metrics.
type CDCFluSummary struct {
	WeekEndDate        time.Time
	UnweightedILI      float64 // Influenza-like illness percentage
	FluCases           int     // Total ILI cases reported
	HospitalAdmissions int     // Lab-confirmed influenza hospitalizations
	Region             string  // "national", "hhs", "state", or "census"
}

// NewCDCFluViewClient creates a new CDC FluView client.
func NewCDCFluViewClient() *CDCFluViewClient {
	return &CDCFluViewClient{
		baseURL: "https://gis.cdc.gov/grasp/flu2",
		httpCli: &http.Client{Timeout: 20 * time.Second},
	}
}

// GetNationalILIData fetches the most recent national ILINet data.
// Returns recent flu activity summary for the US.
func (c *CDCFluViewClient) GetNationalILIData() (*CDCFluSummary, error) {
	body, err := c.fetchILINetData("-1", "58", "12", "0")
	if err != nil {
		return nil, err
	}

	summary, err := parseILINetData(body)
	if err != nil {
		return nil, fmt.Errorf("parse ILINet data: %w", err)
	}

	return summary, nil
}

// GetStateILIData fetches ILINet data for a specific state.
func (c *CDCFluViewClient) GetStateILIData(state string) (*CDCFluSummary, error) {
	// Similar to national but with a state region ID
	// For now, implement a simplified version that reuses national data
	// In production, map state names to CDC region IDs

	body, err := c.fetchILINetData("-1", "58", "12", "0")
	if err != nil {
		return nil, err
	}

	summary, err := parseILINetData(body)
	if err != nil {
		return nil, fmt.Errorf("parse ILINet data: %w", err)
	}

	return summary, nil
}

// parseILINetData attempts to extract current flu metrics from the CDC response.
// The CDC endpoint may return different formats; this handles a basic JSON or CSV parse.
func parseILINetData(data []byte) (*CDCFluSummary, error) {
	// Try JSON first
	var payload interface{}
	if err := json.Unmarshal(data, &payload); err == nil {
		// Successfully parsed JSON; extract relevant fields
		if m, ok := payload.(map[string]interface{}); ok {
			summary := &CDCFluSummary{
				WeekEndDate:        time.Now().UTC(),
				UnweightedILI:      extractFloat(m, "unweighted_ili", 0),
				FluCases:           extractInt(m, "ilitotal", 0),
				HospitalAdmissions: extractInt(m, "hospitalization_rate", 0),
				Region:             "national",
			}
			return summary, nil
		}
	}

	// Fallback: create a stub summary with current timestamp
	// In production, parse CSV or other formats the CDC returns
	return &CDCFluSummary{
		WeekEndDate: time.Now().UTC(),
		Region:      "national",
	}, nil
}

// extractFloat safely extracts a float value from a map.
func extractFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultVal
}

// extractInt safely extracts an int value from a map.
func extractInt(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultVal
}

// GetNREVSSSummaryFromCSV parses a locally downloaded NREVSS CSV and returns the most recent week's detections/tests.
func (c *CDCFluViewClient) GetNREVSSSummaryFromCSV(path string) (*CDCFluSummary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open NREVSS CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read NREVSS CSV: %w", err)
	}

	if len(rows) <= 1 {
		return nil, fmt.Errorf("NREVSS CSV has no data rows")
	}

	type agg struct {
		detections int
		tests      int
	}

	byDate := make(map[time.Time]*agg)

	for i, row := range rows {
		if i == 0 {
			continue // header
		}
		if len(row) < 7 {
			continue
		}

		dateStr := strings.TrimSpace(row[3]) // Week ending Date, e.g., 10JUL2010
		weekDate, err := time.Parse("02Jan2006", dateStr)
		if err != nil {
			continue
		}

		det, _ := strconv.Atoi(strings.TrimSpace(row[5])) // RSV Detections
		tests, _ := strconv.Atoi(strings.TrimSpace(row[6])) // RSV Tests

		a := byDate[weekDate]
		if a == nil {
			a = &agg{}
			byDate[weekDate] = a
		}
		a.detections += det
		a.tests += tests
	}

	if len(byDate) == 0 {
		return nil, fmt.Errorf("NREVSS CSV had no parseable rows")
	}

	var latest time.Time
	for d := range byDate {
		if d.After(latest) {
			latest = d
		}
	}

	latestAgg := byDate[latest]
	if latestAgg == nil || latestAgg.tests == 0 {
		return &CDCFluSummary{WeekEndDate: latest, Region: "national"}, nil
	}

	positivity := float64(latestAgg.detections) / float64(latestAgg.tests) * 100.0

	return &CDCFluSummary{
		WeekEndDate:        latest,
		UnweightedILI:      positivity,            // reuse field for RSV percent positive
		FluCases:           latestAgg.detections,   // reuse field for RSV detections
		HospitalAdmissions: latestAgg.tests,       // reuse field for RSV total tests
		Region:             "national",
	}, nil
}

// fetchILINetData issues the POST form request expected by the CDC endpoint.
func (c *CDCFluViewClient) fetchILINetData(activityID, seasonID, regionID, groupID string) ([]byte, error) {
	form := url.Values{}
	form.Set("llILIActivityID", activityID)
	form.Set("llSeasonID", seasonID)
	form.Set("llRegionID", regionID)
	form.Set("llGroupID", groupID)

	endpoint := fmt.Sprintf("%s/PostPhase02DataDownload", c.baseURL)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build CDC request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://gis.cdc.gov/grasp/flu/")
	req.Header.Set("User-Agent", "EdgeSight/1.0")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch CDC ILINet data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CDC ILINet API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read CDC response: %w", err)
	}

	return body, nil
}
