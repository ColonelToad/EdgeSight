package clients

// CoinGeckoClient handles interactions with the CoinGecko API
type CoinGeckoClient struct {
	baseURL string
}

// NewCoinGeckoClient creates a new CoinGecko API client
func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		baseURL: "https://api.coingecko.com/api/v3",
	}
}
