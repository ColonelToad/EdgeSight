package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CityBikesClient handles interactions with the CityBikes API
type CityBikesClient struct {
	baseURL string
	httpCli *http.Client
}

// NewCityBikesClient creates a new CityBikes API client
func NewCityBikesClient() *CityBikesClient {
	return &CityBikesClient{
		baseURL: "http://api.citybik.es/v2",
		httpCli: &http.Client{Timeout: 10 * time.Second},
	}
}

// NetworksResponse is a subset of the /v2/networks response.
type NetworksResponse struct {
    Networks []Network `json:"networks"`
}

// Network holds brief network metadata.
type Network struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Location BikeLocation `json:"location"`
    Href     string   `json:"href"`
}

// Location holds geographical information about a network.
type BikeLocation struct {
    City      string  `json:"city"`
    Country   string  `json:"country"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}

// ListNetworks fetches the bike networks catalogue.
func (c *CityBikesClient) ListNetworks() (*NetworksResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/networks", c.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var parsed NetworksResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &parsed, nil
}

