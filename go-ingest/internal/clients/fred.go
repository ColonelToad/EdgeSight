package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// FREDClient fetches economic time series from the St. Louis Fed (FRED).
// Docs: https://fred.stlouisfed.org/docs/api/fred/series_observations.html
// Free tier requires API key via env.
type FREDClient struct {
	apiKey  string
	httpCli *http.Client
}

// NewFREDClient creates a new FRED client.
func NewFREDClient(apiKey string) *FREDClient {
	return &FREDClient{
		apiKey:  apiKey,
		httpCli: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetNasdaqComposite returns the latest NASDAQ Composite close via FRED series NASDAQCOM.
func (c *FREDClient) GetNasdaqComposite() (*NASDAQMarketSummary, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("FRED API key required")
	}

	// NASDAQ Composite series_id: NASDAQCOM (daily)
	url := fmt.Sprintf("https://api.stlouisfed.org/fred/series/observations?series_id=NASDAQCOM&api_key=%s&file_type=json&sort_order=desc&limit=1", c.apiKey)

	resp, err := c.httpCli.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch FRED NASDAQ: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FRED NASDAQ returned %d", resp.StatusCode)
	}

	var payload struct {
		Observations []struct {
			Value string `json:"value"`
		} `json:"observations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode FRED NASDAQ: %w", err)
	}

	if len(payload.Observations) == 0 {
		return nil, fmt.Errorf("no observations from FRED NASDAQ")
	}

	valStr := payload.Observations[0].Value
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return nil, fmt.Errorf("parse NASDAQ value: %w", err)
	}

	return &NASDAQMarketSummary{IndexValue: val, VolumeTraded: 0}, nil
}
