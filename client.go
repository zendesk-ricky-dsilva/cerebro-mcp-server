package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// CerebroClient handles communication with the Cerebro API
type CerebroClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewCerebroClient creates a new Cerebro API client
func NewCerebroClient(baseURL, token string) *CerebroClient {
	return &CerebroClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      token,
	}
}

// buildURL builds the API URL with the given parameters
func (c *CerebroClient) buildURL(params CerebroAPIParameters) string {
	urlParams := url.Values{}
	urlParams.Set(fmt.Sprintf("search[%s]", params.searchKey), params.searchValue)

	if len(params.includes) != 0 {
		urlParams.Add("includes", params.includes)
	}

	if len(params.inlines) != 0 {
		urlParams.Add("inlines", params.inlines)
	}

	return fmt.Sprintf("%s?%s", c.baseURL, urlParams.Encode())
}

// makeRequest makes an authenticated HTTP request to the Cerebro API
func (c *CerebroClient) makeRequest(ctx context.Context, apiURL string) (*APIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &apiResponse, nil
}
