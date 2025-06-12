package services

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/farisdewantoro/golang-day-trading-signal/models"
)

// YahooFinanceService handles fetching stock data from Yahoo Finance
type YahooFinanceService struct {
	client *http.Client
}

// NewYahooFinanceService creates a new Yahoo Finance service
func NewYahooFinanceService() *YahooFinanceService {
	return &YahooFinanceService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchOHLCData fetches OHLC data for a given stock symbol
func (y *YahooFinanceService) FetchOHLCData(symbol string) ([]models.OHLCData, error) {
	baseURL := "https://query1.finance.yahoo.com/v8/finance/chart/"
	params := url.Values{}
	params.Add("interval", "5m")
	params.Add("range", "2d")

	url := fmt.Sprintf("%s%s?%s", baseURL, symbol+".JK", params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://finance.yahoo.com/")

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from Yahoo Finance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Yahoo Finance API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle gzip compression
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(io.NopCloser(io.NewSectionReader(bytes.NewReader(body), 0, int64(len(body)))))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()

		body, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress gzip response: %w", err)
		}
	}

	var yahooResp models.YahooFinanceResponse
	if err := json.Unmarshal(body, &yahooResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Yahoo Finance response: %w", err)
	}

	if len(yahooResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for symbol: %s", symbol)
	}

	result := yahooResp.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data available for symbol: %s", symbol)
	}

	quote := result.Indicators.Quote[0]
	ohlcData := make([]models.OHLCData, 0)

	for i, timestamp := range result.Timestamp {
		// Skip if any required data is missing
		if i >= len(quote.Open) || i >= len(quote.High) || i >= len(quote.Low) ||
			i >= len(quote.Close) || i >= len(quote.Volume) {
			continue
		}

		// Skip if any value is 0 (missing data)
		if quote.Open[i] == 0 || quote.High[i] == 0 || quote.Low[i] == 0 || quote.Close[i] == 0 {
			continue
		}

		data := models.OHLCData{
			Timestamp: time.Unix(timestamp, 0),
			Open:      quote.Open[i],
			High:      quote.High[i],
			Low:       quote.Low[i],
			Close:     quote.Close[i],
			Volume:    quote.Volume[i],
		}
		ohlcData = append(ohlcData, data)
	}

	if len(ohlcData) == 0 {
		return nil, fmt.Errorf("no valid OHLC data found for symbol: %s", symbol)
	}

	return ohlcData, nil
}
