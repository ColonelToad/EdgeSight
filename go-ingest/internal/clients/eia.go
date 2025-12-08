package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EIAClient queries the US Energy Information Administration API
// EIA provides comprehensive energy data including generation, consumption, and prices
// API Docs: https://www.eia.gov/opendata/
type EIAClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// EIAEnergySummary represents aggregated energy generation and price data
type EIAEnergySummary struct {
	ElectricityGenerationMWh float64 // Total electricity generation
	NaturalGasPriceMmbtu     float64 // Natural gas spot price ($/MMBtu)
	CoalPriceTon             float64 // Coal price ($/short ton)
	RenewableGenerationMWh   float64 // Renewable electricity generation
	TotalDemandMWh           float64 // Total electricity demand
}

// EIAResponse represents the API response structure
type EIAResponse struct {
	Response struct {
		Data []struct {
			Period string  `json:"period"`
			Value  float64 `json:"value"`
		} `json:"data"`
	} `json:"response"`
}

// NewEIAClient creates a new EIA API client
func NewEIAClient(apiKey string) *EIAClient {
	return &EIAClient{
		APIKey:  apiKey,
		BaseURL: "https://api.eia.gov/v2",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetElectricityGeneration fetches total US electricity generation data
func (c *EIAClient) GetElectricityGeneration() (*EIAEnergySummary, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("EIA API key required")
	}

	// Query electricity net generation (thousand megawatthours)
	// Series ID: ELEC.GEN.ALL-US-99.M (monthly total generation)
	endpoint := fmt.Sprintf("/electricity/electric-power-operational-data/data/?api_key=%s&frequency=monthly&data[0]=generation&facets[location][]=US&sort[0][column]=period&sort[0][direction]=desc&offset=0&length=1", c.APIKey)

	data, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var resp EIAResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(resp.Response.Data) == 0 {
		return nil, fmt.Errorf("no data returned from EIA")
	}

	// Convert thousand MWh to MWh
	generationMWh := resp.Response.Data[0].Value * 1000

	// For demo purposes, return mock data for other fields
	// In production, you'd make additional API calls for each metric
	return &EIAEnergySummary{
		ElectricityGenerationMWh: generationMWh,
		NaturalGasPriceMmbtu:     2.85,
		CoalPriceTon:             52.30,
		RenewableGenerationMWh:   generationMWh * 0.21, // ~21% renewable
		TotalDemandMWh:           generationMWh * 0.98, // Assume 2% losses
	}, nil
}

// GetNaturalGasPrice fetches current natural gas spot prices
func (c *EIAClient) GetNaturalGasPrice() (float64, error) {
	if c.APIKey == "" {
		return 0, fmt.Errorf("EIA API key required")
	}

	// Query Henry Hub Natural Gas Spot Price
	// Series ID: NG.RNGWHHD.D
	endpoint := fmt.Sprintf("/natural-gas/pri/spt/data/?api_key=%s&frequency=daily&data[0]=value&facets[series][]=RNGWHHD&sort[0][column]=period&sort[0][direction]=desc&offset=0&length=1", c.APIKey)

	data, err := c.makeRequest(endpoint)
	if err != nil {
		return 0, err
	}

	var resp EIAResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	if len(resp.Response.Data) == 0 {
		return 0, fmt.Errorf("no gas price data returned")
	}

	return resp.Response.Data[0].Value, nil
}

// GetEnergySummary fetches comprehensive energy data
func (c *EIAClient) GetEnergySummary() (*EIAEnergySummary, error) {
	summary, err := c.GetElectricityGeneration()
	if err != nil {
		return nil, err
	}

	// Try to get natural gas price, but don't fail if it errors
	gasPrice, err := c.GetNaturalGasPrice()
	if err == nil {
		summary.NaturalGasPriceMmbtu = gasPrice
	}

	return summary, nil
}

// makeRequest makes an HTTP request to the EIA API
func (c *EIAClient) makeRequest(endpoint string) ([]byte, error) {
	url := c.BaseURL + endpoint

	var lastErr error
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		resp, err := c.Client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("EIA API returned status %d", resp.StatusCode)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var result json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			lastErr = fmt.Errorf("decode response: %w", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return result, nil
	}

	return nil, lastErr
}
