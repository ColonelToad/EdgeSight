package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// OpenMeteoClient handles interactions with the Open-Meteo API.
type OpenMeteoClient struct {
	baseURL string
	httpCli *http.Client
}

// NewOpenMeteoClient creates a new Open-Meteo API client
func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		baseURL: "https://api.open-meteo.com/v1",
		httpCli: &http.Client{Timeout: 10 * time.Second},
	}
}

// CurrentWeatherResponse represents a subset of the forecast current weather response.
type CurrentWeatherResponse struct {
	Latitude  float64      `json:"latitude"`
	Longitude float64      `json:"longitude"`
	Current   CurrentBlock `json:"current"`
}

// CurrentBlock holds the current weather metrics requested.
type CurrentBlock struct {
	Time             string  `json:"time"`
	Temperature2m    float64 `json:"temperature_2m"`
	WindSpeed10m     float64 `json:"wind_speed_10m"`
	RelativeHumidity float64 `json:"relative_humidity_2m"`
}

// GetCurrentWeather fetches current weather for provided coordinates.
func (c *OpenMeteoClient) GetCurrentWeather(lat, lon float64) (*CurrentWeatherResponse, error) {
	q := url.Values{}
	q.Set("latitude", fmt.Sprintf("%f", lat))
	q.Set("longitude", fmt.Sprintf("%f", lon))
	q.Set("current", "temperature_2m,wind_speed_10m,relative_humidity_2m")

	reqURL := fmt.Sprintf("%s/forecast?%s", c.baseURL, q.Encode())
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
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

	var parsed CurrentWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &parsed, nil
}
