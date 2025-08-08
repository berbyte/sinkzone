package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetQueries() ([]DNSQuery, error) {
	resp, err := c.client.Get(c.baseURL + "/api/queries")
	if err != nil {
		return nil, fmt.Errorf("failed to get queries: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't return it since we're already returning
			fmt.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var queries []DNSQuery
	if err := json.NewDecoder(resp.Body).Decode(&queries); err != nil {
		return nil, fmt.Errorf("failed to decode queries: %w", err)
	}

	return queries, nil
}

func (c *Client) GetFocusMode() (*FocusModeState, error) {
	resp, err := c.client.Get(c.baseURL + "/api/focus")
	if err != nil {
		return nil, fmt.Errorf("failed to get focus mode: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't return it since we're already returning
			fmt.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var state FocusModeState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode focus mode: %w", err)
	}

	return &state, nil
}

func (c *Client) SetFocusMode(enabled bool, duration string) error {
	req := struct {
		Enabled  bool   `json:"enabled"`
		Duration string `json:"duration,omitempty"`
	}{
		Enabled:  enabled,
		Duration: duration,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/focus", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to set focus mode: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't return it since we're already returning
			fmt.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetState() (*ResolverState, error) {
	resp, err := c.client.Get(c.baseURL + "/api/state")
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't return it since we're already returning
			fmt.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var state ResolverState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode state: %w", err)
	}

	return &state, nil
}

func (c *Client) HealthCheck() error {
	// log.Printf("API Client: Attempting health check to %s/health", c.baseURL)

	resp, err := c.client.Get(c.baseURL + "/health")
	if err != nil {
		log.Printf("API Client: Health check failed with error: %v", err)
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't return it since we're already returning
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	// log.Printf("API Client: Health check response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("API Client: Health check failed with status %d, body: %s", resp.StatusCode, string(body))
		return fmt.Errorf("health check returned status: %d", resp.StatusCode)
	}

	// body, _ := io.ReadAll(resp.Body)
	// log.Printf("API Client: Health check successful, response: %s", string(body))
	return nil
}
