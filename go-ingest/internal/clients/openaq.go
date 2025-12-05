package clients

import (
	"bytes"
	"io"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// OpenAQClient handles interactions with the OpenAQ API
type OpenAQClient struct {
	baseURL string
	apiKey  string
	httpCli *http.Client
}

// NewOpenAQClient creates a new OpenAQ API client
func NewOpenAQClient(apiKey string) *OpenAQClient {
    return &OpenAQClient{
        baseURL: "https://api.openaq.org/v3",
        apiKey:  apiKey,
        httpCli: &http.Client{Timeout: 15 * time.Second},
    }
}

// LocationsResponse represents the response from /v3/locations
type LocationsResponse struct {
    Meta    ResponseMeta   `json:"meta"`
    Results []OpenAQLocation `json:"results"`  // Changed here
}

// OpenAQLocation represents a monitoring location (renamed from Location)
type OpenAQLocation struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`
    Locality    string      `json:"locality"`
    Timezone    string      `json:"timezone"`
    Country     Country     `json:"country"`
    Owner       Owner       `json:"owner"`
    Provider    Provider    `json:"provider"`
    IsMobile    bool        `json:"isMobile"`
    IsMonitor   bool        `json:"isMonitor"`
    Coordinates Coordinates `json:"coordinates"`
	DatetimeLast *DatetimeInfo `json:"datetimeLast"`
}


// Country represents country information
type Country struct {
    ID   int    `json:"id"`
    Code string `json:"code"`
    Name string `json:"name"`
}

// Owner represents the owner of a location
type Owner struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// Provider represents the data provider
type Provider struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// Coordinates represents geographic coordinates
type Coordinates struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}

// LatestResponse represents the response from /v3/locations/{id}/latest
type LatestResponse struct {
    Meta    ResponseMeta     `json:"meta"`
    Results []LatestMeasurement `json:"results"`
}

// LatestMeasurement represents a single latest measurement
type LatestMeasurement struct {
    Datetime    DatetimeInfo `json:"datetime"`
    Value       float64      `json:"value"`
    Coordinates Coordinates  `json:"coordinates"`
    Parameter   Parameter    `json:"parameter"`
}

// DatetimeInfo contains UTC and local timestamps
type DatetimeInfo struct {
    UTC   string `json:"utc"`
    Local string `json:"local"`
}

// Parameter represents a measurement parameter
type Parameter struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Units       string `json:"units"`
    DisplayName string `json:"displayName"`
}

// ResponseMeta contains metadata about the response
type ResponseMeta struct {
    Name       string `json:"name"`
    License    string `json:"license"`
    Website    string `json:"website"`
    Page       int    `json:"page"`
    Limit      int    `json:"limit"`
    Found      interface{} `json:"found"`
}

// SensorsResponse represents the response from /v3/sensors
type SensorsResponse struct {
    Meta    ResponseMeta `json:"meta"`
    Results []Sensor     `json:"results"`
}

// Sensor combines Metadata (what is it?) with Latest Data (what is the value?)
type Sensor struct {
    ID        int           `json:"id"`
    Name      string        `json:"name"`      // e.g. "PurpleAir-Primary"
    Parameter Parameter     `json:"parameter"` // Contains DisplayName & Units
    Latest    SensorReading `json:"latest"`
}

// SensorReading is the actual data point inside a sensor
type SensorReading struct {
    Value    float64      `json:"value"`
    Datetime DatetimeInfo `json:"datetime"`
}

func (c *OpenAQClient) GetSensorsByLocationID(locationID int) (*SensorsResponse, error) {
    // Correct endpoint: /v3/locations/{id}/sensors (not /v3/sensors)
    reqURL := fmt.Sprintf("%s/locations/%d/sensors", c.baseURL, locationID)
    
    req, err := http.NewRequest(http.MethodGet, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }
    if c.apiKey != "" {
        req.Header.Set("X-API-Key", c.apiKey)
    }

    resp, err := c.httpCli.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
    }

    var parsed SensorsResponse
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &parsed, nil
}

// GetLocationsByCity fetches locations in a city
func (c *OpenAQClient) GetLocationsByCity(city string, limit int) (*LocationsResponse, error) {
    if c.apiKey == "" {
        return nil, fmt.Errorf("openaq api key is required")
    }

    q := url.Values{}
    q.Set("city", city)
    q.Set("limit", fmt.Sprintf("%d", limit))

    reqURL := fmt.Sprintf("%s/locations?%s", c.baseURL, q.Encode())
    req, err := http.NewRequest(http.MethodGet, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }
    req.Header.Set("X-API-Key", c.apiKey)

    resp, err := c.httpCli.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
    }

    var parsed LocationsResponse
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &parsed, nil
}

// GetLatestByLocationID fetches latest measurements for a specific location
func (c *OpenAQClient) GetLatestByLocationID(locationID int) (*LatestResponse, error) {
    if c.apiKey == "" {
        return nil, fmt.Errorf("openaq api key is required")
    }

    reqURL := fmt.Sprintf("%s/locations/%d/latest", c.baseURL, locationID)
    req, err := http.NewRequest(http.MethodGet, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }
    req.Header.Set("X-API-Key", c.apiKey)

    resp, err := c.httpCli.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
    }

    // --- START X-RAY CODE ---
    // Read the raw body into a byte array
    bodyBytes, _ := io.ReadAll(resp.Body)
    
    // Print it to the console so we can see the TRUTH
    fmt.Println("DEBUG RAW JSON RESPONSE:", string(bodyBytes))

    // Restore the body so the JSON decoder can read it again
    resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    // --- END X-RAY CODE ---

    var parsed LatestResponse
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &parsed, nil
}

// GetLocationsByCoordinates fetches locations near a coordinate point
func (c *OpenAQClient) GetLocationsByCoordinates(lat, lon float64, radius int, limit int) (*LocationsResponse, error) {
    if c.apiKey == "" {
        return nil, fmt.Errorf("openaq api key is required")
    }

    q := url.Values{}
    q.Set("coordinates", fmt.Sprintf("%f,%f", lat, lon))
    q.Set("radius", fmt.Sprintf("%d", radius)) // radius in meters
    q.Set("limit", fmt.Sprintf("%d", limit))

    reqURL := fmt.Sprintf("%s/locations?%s", c.baseURL, q.Encode())
    req, err := http.NewRequest(http.MethodGet, reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }
    req.Header.Set("X-API-Key", c.apiKey)

    resp, err := c.httpCli.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
    }

    var parsed LocationsResponse
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &parsed, nil
}