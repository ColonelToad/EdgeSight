package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// NASSClient queries the USDA National Agricultural Statistics Service API
// NASS provides crop data, livestock statistics, and agricultural economics
// API Docs: https://quickstats.nass.usda.gov/api
type NASSClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// NASSCropSummary represents aggregated crop statistics
type NASSCropSummary struct {
	CropType          string  // e.g., "CORN", "SOYBEANS", "WHEAT"
	ProductionBushels float64 // Total production in bushels
	YieldPerAcre      float64 // Yield (bushels/acre)
	HarvestedAcres    float64 // Acres harvested
	PricePerBushel    float64 // Price received ($/bushel)
	State             string  // State abbreviation
	Year              int     // Year of data
}

// NASSResponse represents the NASS API response
type NASSResponse struct {
	Data []struct {
		CommodityDesc string      `json:"commodity_desc"`
		Year          json.Number `json:"year"`
		State         string      `json:"state_alpha"`
		Value         string      `json:"Value"`
		Unit          string      `json:"unit_desc"`
		StatisticCat  string      `json:"statisticcat_desc"`
	} `json:"data"`
}

// NewNASSClient creates a new NASS API client
func NewNASSClient(apiKey string) *NASSClient {
	return &NASSClient{
		APIKey:  apiKey,
		BaseURL: "https://quickstats.nass.usda.gov/api",
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// GetCropProduction fetches production data for a specific crop and state
func (c *NASSClient) GetCropProduction(crop, state string, year int) (*NASSCropSummary, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("NASS API key required")
	}

	// Build query parameters
	params := url.Values{}
	params.Set("key", c.APIKey)
	params.Set("commodity_desc", crop)
	params.Set("year", fmt.Sprintf("%d", year))
	params.Set("state_alpha", state)
	params.Set("statisticcat_desc", "PRODUCTION")
	params.Set("format", "JSON")

	endpoint := fmt.Sprintf("/api_GET/?%s", params.Encode())

	data, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var resp NASSResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no data found for %s in %s (%d)", crop, state, year)
	}

	row := resp.Data[0]
	prodVal, _ := parseNumber(row.Value)
	parsedYear, _ := row.Year.Int64()

	summary := &NASSCropSummary{
		CropType:          crop,
		State:             state,
		Year:              int(parsedYear),
		ProductionBushels: prodVal,
		YieldPerAcre:      0,
		HarvestedAcres:    0,
		PricePerBushel:    0,
	}

	return summary, nil
}

// GetNationalCropSummary fetches aggregated national crop data
func (c *NASSClient) GetNationalCropSummary(crop string) (*NASSCropSummary, error) {
	currentYear := time.Now().Year() - 1 // Use previous year for complete data

	// Get national (US-level) data
	return c.GetCropProduction(crop, "US", currentYear)
}

// GetStateCropSummary fetches state-level crop data
func (c *NASSClient) GetStateCropSummary(crop, state string) (*NASSCropSummary, error) {
	currentYear := time.Now().Year() - 1

	return c.GetCropProduction(crop, state, currentYear)
}

// makeRequest makes an HTTP request to the NASS API
func (c *NASSClient) makeRequest(endpoint string) ([]byte, error) {
	url := c.BaseURL + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NASS API returned status %d", resp.StatusCode)
	}

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

// parseNumber converts string with optional commas into float64.
func parseNumber(val string) (float64, error) {
	clean := ""
	for i := 0; i < len(val); i++ {
		if val[i] == ',' {
			continue
		}
		clean += string(val[i])
	}
	return strconv.ParseFloat(clean, 64)
}
