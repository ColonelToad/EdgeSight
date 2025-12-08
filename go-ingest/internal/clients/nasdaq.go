package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NASDAQClient fetches market data from NASDAQ Data Link (formerly Quandl).
// Requires a free API key from https://data.nasdaq.com
type NASDAQClient struct {
	baseURL string
	apiKey  string
	httpCli *http.Client
}

// NASDAQMarketSummary aggregates current market metrics.
type NASDAQMarketSummary struct {
	IndexValue       float64 // NASDAQ Composite index value
	VolumeTraded     int64   // Total shares traded
	AdvancingStocks  int     // Number of stocks advancing
	DecliningStocks  int     // Number of stocks declining
	MarketCapBillions float64 // Total market cap in billions USD
}

// NewNASDAQClient creates a NASDAQ Data Link client.
func NewNASDAQClient(apiKey string) *NASDAQClient {
	return &NASDAQClient{
		baseURL: "https://data.nasdaq.com/api/v3",
		apiKey:  apiKey,
		httpCli: &http.Client{Timeout: 20 * time.Second},
	}
}

// GetMarketSummary fetches current NASDAQ composite index and market metrics.
func (c *NASDAQClient) GetMarketSummary() (*NASDAQMarketSummary, error) {
	// NASDAQ Data Link endpoint for composite index
	// Example: /datasets/NASDAQOMX/COMP.json?api_key=XXX&limit=1
	url := fmt.Sprintf("%s/datasets/NASDAQOMX/COMP.json?api_key=%s&limit=1", c.baseURL, c.apiKey)

	resp, err := c.httpCli.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch NASDAQ data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NASDAQ API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read NASDAQ response: %w", err)
	}

	summary, err := parseMarketData(body)
	if err != nil {
		return nil, fmt.Errorf("parse NASDAQ data: %w", err)
	}

	return summary, nil
}

// parseMarketData extracts market metrics from NASDAQ Data Link response.
func parseMarketData(data []byte) (*NASDAQMarketSummary, error) {
	var payload struct {
		Dataset struct {
			Data [][]interface{} `json:"data"`
		} `json:"dataset"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	if len(payload.Dataset.Data) == 0 {
		return nil, fmt.Errorf("no data returned from NASDAQ")
	}

	// NASDAQ Data Link format: [date, index_value, high, low, volume, ...]
	row := payload.Dataset.Data[0]
	if len(row) < 2 {
		return nil, fmt.Errorf("invalid data format")
	}

	indexValue := 0.0
	if val, ok := row[1].(float64); ok {
		indexValue = val
	}

	volume := int64(0)
	if len(row) > 4 {
		if val, ok := row[4].(float64); ok {
			volume = int64(val)
		}
	}

	return &NASDAQMarketSummary{
		IndexValue:   indexValue,
		VolumeTraded: volume,
		// Other metrics would require additional API calls or datasets
	}, nil
}
