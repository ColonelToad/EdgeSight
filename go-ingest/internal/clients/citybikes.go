package clients

// CityBikesClient handles interactions with the CityBikes API
type CityBikesClient struct {
	baseURL string
}

// NewCityBikesClient creates a new CityBikes API client
func NewCityBikesClient() *CityBikesClient {
	return &CityBikesClient{
		baseURL: "https://api.citybik.es/v2",
	}
}
