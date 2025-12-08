package clients

import (
	"fmt"
	"math/rand"
	"time"
)

// GridClient queries grid status and load data
// This is a mock client that simulates grid monitoring data
// In production, this would integrate with ISOs like CAISO, PJM, ERCOT, etc.
type GridClient struct {
	Region string
}

// GridStatus represents current power grid conditions
type GridStatus struct {
	LoadMW             float64 // Current grid load in megawatts
	CapacityMW         float64 // Total available capacity in megawatts
	UtilizationPercent float64 // Load as percentage of capacity
	FrequencyHz        float64 // Grid frequency (should be ~60Hz in US, ~50Hz in Europe)
	Status             string  // "Normal", "Alert", "Emergency"
	RenewablesMW       float64 // Current renewable generation in MW
}

// NewGridClient creates a new grid monitoring client
func NewGridClient(region string) *GridClient {
	return &GridClient{
		Region: region,
	}
}

// GetGridStatus fetches current grid status and load
// This is a mock implementation that generates realistic data
func (c *GridClient) GetGridStatus() (*GridStatus, error) {
	// Seed randomizer for realistic variation
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Base load varies by time of day (mock implementation)
	hour := time.Now().Hour()
	baseLoad := 25000.0 // MW
	
	// Peak load during afternoon/evening (2pm - 8pm)
	if hour >= 14 && hour <= 20 {
		baseLoad = 35000.0
	} else if hour >= 6 && hour < 14 {
		baseLoad = 30000.0
	} else {
		baseLoad = 22000.0 // Night/early morning
	}

	// Add some randomness (±10%)
	loadMW := baseLoad * (0.9 + r.Float64()*0.2)
	capacityMW := 45000.0
	utilizationPercent := (loadMW / capacityMW) * 100

	// Frequency should be close to 60Hz (US grid)
	frequencyHz := 59.95 + r.Float64()*0.1

	// Renewables vary by time (solar peak during day)
	renewablesMW := 5000.0
	if hour >= 9 && hour <= 16 {
		renewablesMW = 8000.0 + r.Float64()*2000.0 // High solar during day
	} else if hour >= 17 && hour <= 22 {
		renewablesMW = 6000.0 + r.Float64()*1000.0 // Wind picks up evening
	} else {
		renewablesMW = 3000.0 + r.Float64()*1000.0 // Mostly wind at night
	}

	// Determine status based on utilization
	status := "Normal"
	if utilizationPercent > 90 {
		status = "Emergency"
	} else if utilizationPercent > 80 {
		status = "Alert"
	}

	return &GridStatus{
		LoadMW:             loadMW,
		CapacityMW:         capacityMW,
		UtilizationPercent: utilizationPercent,
		FrequencyHz:        frequencyHz,
		Status:             status,
		RenewablesMW:       renewablesMW,
	}, nil
}

// GetRegionalLoad fetches load data for a specific region
// In production, this would query ISO-specific APIs (CAISO, ERCOT, PJM, etc.)
func (c *GridClient) GetRegionalLoad(region string) (float64, error) {
	// Mock regional load data
	regionalLoads := map[string]float64{
		"CAISO":    25000.0, // California
		"ERCOT":    45000.0, // Texas
		"PJM":      85000.0, // Mid-Atlantic
		"NYISO":    20000.0, // New York
		"ISO-NE":   15000.0, // New England
		"MISO":     75000.0, // Midwest
		"SPP":      35000.0, // Southwest Power Pool
	}

	if load, ok := regionalLoads[region]; ok {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		// Add ±5% variation
		return load * (0.95 + r.Float64()*0.1), nil
	}

	return 0, fmt.Errorf("region not supported: %s", region)
}
