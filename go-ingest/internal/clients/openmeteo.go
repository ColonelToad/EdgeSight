package clients

// OpenMeteoClient handles interactions with the Open-Meteo API
type OpenMeteoClient struct {
	baseURL string
}

// NewOpenMeteoClient creates a new Open-Meteo API client
func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		baseURL: "https://api.open-meteo.com/v1",
	}
}
