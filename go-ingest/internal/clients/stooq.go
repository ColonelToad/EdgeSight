package clients

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// StooqClient fetches free market data from stooq.pl as a lightweight fallback.
type StooqClient struct {
	baseURL string
	httpCli *http.Client
}

// NewStooqClient creates a Stooq client with sensible defaults.
func NewStooqClient() *StooqClient {
	return &StooqClient{
		baseURL: "https://stooq.pl/q/l/",
		httpCli: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetNasdaqComposite returns NASDAQ composite via Stooq (^ndq), mapped into NASDAQMarketSummary.
// Stooq CSV format: Symbol,Date,Time,Open,High,Low,Close,Volume
func (c *StooqClient) GetNasdaqComposite() (*NASDAQMarketSummary, error) {
	// i=d gives daily; f=sd2t2ohlcv includes symbol/date/time/ohlcv; h&e=csv ensures headers and CSV
	url := fmt.Sprintf("%s?s=^ndq&f=sd2t2ohlcv&h&e=csv", c.baseURL)

	resp, err := c.httpCli.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch Stooq NASDAQ: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Stooq NASDAQ returned %d: %s", resp.StatusCode, string(body))
	}

	reader := csv.NewReader(resp.Body)
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse Stooq CSV: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("Stooq NASDAQ CSV missing data rows")
	}

	row := rows[1]
	if len(row) < 8 {
		return nil, fmt.Errorf("Stooq NASDAQ CSV malformed")
	}

	closeVal := parseFloatSafe(row[6])
	vol := parseInt64Safe(row[7])

	return &NASDAQMarketSummary{
		IndexValue:   closeVal,
		VolumeTraded: vol,
	}, nil
}

func parseFloatSafe(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func parseInt64Safe(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}
