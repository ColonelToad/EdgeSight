package clients

// OpenAQClient handles interactions with the OpenAQ API
type OpenAQClient struct {
	baseURL string
	apiKey  string
}

// NewOpenAQClient creates a new OpenAQ API client
func NewOpenAQClient(apiKey string) *OpenAQClient {
	return &OpenAQClient{
		baseURL: "https://api.openaq.org/v2",
		apiKey:  apiKey,
	}
}
