package graphql

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Client is a struct that represents the HTTP client used to interact with the GitHub GraphQL API.
type Client struct {
	URL     string
	headers map[string]string
}

// NewClient creates a new GitHub GraphQL API client with the provided token.
func NewClient(url, token string) *Client {
	// Create a new client with that token
	return &Client{
		URL: url,
		headers: map[string]string{
			"Authorization": "Bearer " + token,
			"Content-Type":  "application/json",
		},
	}
}

func (c *Client) WithDefaultHeader(key, val string) *Client {
	c.headers[key] = val
	return c
}

// Post sends a POST request to the GitHub GraphQL API with the provided requestBody.
func (c *Client) Post(requestBody []byte) ([]byte, error) {
	// Create the HTTP POST request
	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers for the request
	for key, val := range c.headers {
		req.Header.Set(key, val)
	}

	// Create a new HTTP client and send the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}
