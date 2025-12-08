package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EmberClient queries the Ember Climate API for carbon intensity and electricity generation data
// Ember provides global electricity data including carbon intensity, generation mix, and renewable percentages
// API Docs: https://ember-climate.org/data-catalogue/
type EmberClient struct {
	BaseURL string
	Client  *http.Client
}

// EmberElectricitySummary represents aggregated electricity generation and carbon intensity data
type EmberElectricitySummary struct {
	CarbonIntensityGCO2KWh float64 // gCO2/kWh
	RenewablePercent       float64 // Percentage of electricity from renewables
	GenerationTWh          float64 // Total electricity generation in TWh
	CoalPercent            float64 // Percentage from coal
	GasPercent             float64 // Percentage from gas
	NuclearPercent         float64 // Percentage from nuclear
}

// EmberDataPoint represents a single data point from the Ember API
type EmberDataPoint struct {
	Year              int     `json:"year"`
	Country           string  `json:"country"`
	CarbonIntensity   float64 `json:"carbon_intensity_gco2_per_kwh"`
	RenewablePercent  float64 `json:"renewable_percent"`
	GenerationTWh     float64 `json:"generation_twh"`
	CoalPercent       float64 `json:"coal_percent"`
	GasPercent        float64 `json:"gas_percent"`
	NuclearPercent    float64 `json:"nuclear_percent"`
}

// NewEmberClient creates a new Ember API client
func NewEmberClient() *EmberClient {
	return &EmberClient{
		BaseURL: "https://ember-climate.org/app/uploads/2022/07/yearly_full_release.csv",
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetCountrySummary fetches the latest electricity data for a specific country
// Note: Ember provides CSV data files; this is a simplified mock implementation
// In production, you would download and parse the CSV file or use their data API
func (c *EmberClient) GetCountrySummary(countryCode string) (*EmberElectricitySummary, error) {
	// Mock implementation returning realistic data for demonstration
	// In production, this would parse actual Ember CSV data or call their API
	
	// Example mock data for USA
	if countryCode == "USA" || countryCode == "US" {
		return &EmberElectricitySummary{
			CarbonIntensityGCO2KWh: 386.5,
			RenewablePercent:       21.3,
			GenerationTWh:          4178.0,
			CoalPercent:            19.5,
			GasPercent:             38.4,
			NuclearPercent:        18.9,
		}, nil
	}

	// Example mock data for Germany
	if countryCode == "DEU" || countryCode == "DE" || countryCode == "Germany" {
		return &EmberElectricitySummary{
			CarbonIntensityGCO2KWh: 348.2,
			RenewablePercent:       44.6,
			GenerationTWh:          574.5,
			CoalPercent:            29.8,
			GasPercent:             12.6,
			NuclearPercent:        11.4,
		}, nil
	}

	return nil, fmt.Errorf("country data not available for: %s", countryCode)
}

// GetGlobalAverage calculates global average carbon intensity and generation mix
func (c *EmberClient) GetGlobalAverage() (*EmberElectricitySummary, error) {
	// Mock global average data
	return &EmberElectricitySummary{
		CarbonIntensityGCO2KWh: 436.0,
		RenewablePercent:       28.7,
		GenerationTWh:          28466.0,
		CoalPercent:            35.1,
		GasPercent:             23.5,
		NuclearPercent:        9.8,
	}, nil
}

// Internal helper for making HTTP requests (for future real API integration)
func (c *EmberClient) makeRequest(endpoint string) ([]byte, error) {
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
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result []byte
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}
