package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client talks to the Python embedding sidecar.
type Client struct {
	endpoint string
	httpCli  *http.Client
}

// NewClient creates a new embeddings client.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		httpCli:  &http.Client{Timeout: 10 * time.Second},
	}
}

// EmbedRequest represents the payload to the sidecar.
type EmbedRequest struct {
	Text string `json:"text"`
}

// EmbedResponse is the sidecar response.
type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// Embed sends text to the sidecar and returns the vector.
func (c *Client) Embed(text string) ([]float64, error) {
	body, _ := json.Marshal(EmbedRequest{Text: text})
	req, err := http.NewRequest("POST", c.endpoint+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed endpoint returned %d", resp.StatusCode)
	}

	var er EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	return er.Embedding, nil
}
