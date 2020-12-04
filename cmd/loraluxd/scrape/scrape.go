// Package scrape handles scraping the configured LoRaWAN server for lumen sensor
// data points.
package scrape

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Scraper contains the data and receiver methods that scrape a server for lumen
// sensor data points.
type Scraper struct {
	url    string
	client *http.Client
}

// NewScraper returns a new pointer to the Scraper type.
func NewScraper(addr, endpoint string, timeout time.Duration) *Scraper {
	client := http.DefaultClient
	client.Timeout = timeout

	return &Scraper{
		url:    fmt.Sprintf("%s%s", addr, endpoint),
		client: client,
	}
}

// Response is the response expected to receive from a LoRaWAN server to get lumen
// sensor data points from.
type Response struct {
	Data []float64 `json:"data"`
}

// Scrape actually scrapes the LoRaWAN server using the configured Scraper type.
func (s *Scraper) Scrape() (*Response, error) {
	resp, err := s.client.Get(s.url)
	if err != nil {
		return nil, fmt.Errorf("do scrape request: %w", err)
	}
	defer resp.Body.Close()

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	return &respBody, nil
}
