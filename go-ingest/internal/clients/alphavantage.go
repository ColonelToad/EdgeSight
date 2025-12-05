package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// AlphaVantageClient handles interactions with the Alpha Vantage API.
type AlphaVantageClient struct {
	apiKey  string
	baseURL string
	httpCli *http.Client
}

// NewAlphaVantageClient creates a new Alpha Vantage API client.
func NewAlphaVantageClient(apiKey string) *AlphaVantageClient {
	return &AlphaVantageClient{
		apiKey:  apiKey,
		baseURL: "https://www.alphavantage.co/query",
		httpCli: &http.Client{Timeout: 15 * time.Second},
	}
}

// GlobalQuoteResponse represents the Alpha Vantage GLOBAL_QUOTE response.
type GlobalQuoteResponse struct {
	Quote GlobalQuote `json:"Global Quote"`
}

// GlobalQuote holds a minimal subset of quote fields.
type GlobalQuote struct {
	Symbol           string `json:"01. symbol"`
	Open             string `json:"02. open"`
	High             string `json:"03. high"`
	Low              string `json:"04. low"`
	Price            string `json:"05. price"`
	Volume           string `json:"06. volume"`
	LatestTradingDay string `json:"07. latest trading day"`
}

// GetGlobalQuote fetches the latest quote for the given symbol.
func (c *AlphaVantageClient) GetGlobalQuote(symbol string) (*GlobalQuoteResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("alphavantage api key is required")
	}

	q := url.Values{}
	q.Set("function", "GLOBAL_QUOTE")
	q.Set("symbol", symbol)
	q.Set("apikey", c.apiKey)

	reqURL := fmt.Sprintf("%s?%s", c.baseURL, q.Encode())
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

	var parsed GlobalQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &parsed, nil
}
