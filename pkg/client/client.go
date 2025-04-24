// Package client provides functionality for interacting with the Vipps MobilePay ePayment API
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	// TestBaseURL is the base URL for the test environment
	TestBaseURL = "https://apitest.vipps.no"
	// ProductionBaseURL is the base URL for the production environment
	ProductionBaseURL = "https://api.vipps.no"

	// Default timeout for HTTP requests
	defaultTimeout = 30 * time.Second
)

// Client handles communication with the Vipps MobilePay API
type Client struct {
	// HTTP client used for requests
	client *http.Client

	// Base URL for API requests
	BaseURL string

	// API credentials
	ClientID     string
	ClientSecret string
	SubKey       string // Ocp-Apim-Subscription-Key
	MSN          string // Merchant-Serial-Number

	// Access token for API requests
	AccessToken string
	TokenExpiry time.Time

	// System information for HTTP headers
	SystemName          string // Vipps-System-Name
	SystemVersion       string // Vipps-System-Version
	SystemPluginName    string // Vipps-System-Plugin-Name
	SystemPluginVersion string // Vipps-System-Plugin-Version

	// Whether this client is running in test mode
	TestMode bool
}

// NewClient creates a new API client for Vipps MobilePay
func NewClient(clientID, clientSecret, subKey, msn string, testMode bool) *Client {
	baseURL := ProductionBaseURL
	if testMode {
		baseURL = TestBaseURL
	}

	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		SubKey:       subKey,
		MSN:          msn,
		TestMode:     testMode,

		// Default system information
		SystemName:    "go-vipps-mobilepay-sdk",
		SystemVersion: "1.0.0",
	}
}

// SetSystemInfo sets the system information for HTTP headers
func (c *Client) SetSystemInfo(name, version, pluginName, pluginVersion string) {
	if name != "" {
		c.SystemName = name
	}
	if version != "" {
		c.SystemVersion = version
	}
	if pluginName != "" {
		c.SystemPluginName = pluginName
	}
	if pluginVersion != "" {
		c.SystemPluginVersion = pluginVersion
	}
}

// SetTimeout sets the timeout for HTTP requests
func (c *Client) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}

// IsTokenValid checks if the current access token is still valid
func (c *Client) IsTokenValid() bool {
	return c.AccessToken != "" && time.Now().Before(c.TokenExpiry)
}

// GetAccessToken fetches a new access token from the Vipps MobilePay API
func (c *Client) GetAccessToken() error {
	endpoint := "/accesstoken/get"
	url := c.BaseURL + endpoint

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for token request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client_id", c.ClientID)
	req.Header.Set("client_secret", c.ClientSecret)
	req.Header.Set("Ocp-Apim-Subscription-Key", c.SubKey)
	req.Header.Set("Merchant-Serial-Number", c.MSN)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get access token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	c.AccessToken = tokenResp.AccessToken

	// Convert expires_in from string to int
	expiresIn, err := strconv.Atoi(tokenResp.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to convert expires_in to int: %w", err)
	}

	c.TokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return nil
}

// EnsureValidToken makes sure a valid access token is available
func (c *Client) EnsureValidToken() error {
	if !c.IsTokenValid() {
		return c.GetAccessToken()
	}
	return nil
}

// DoRequest performs an HTTP request with the appropriate headers and error handling
func (c *Client) DoRequest(method, endpoint string, body interface{}, idempotencyKey string) ([]byte, int, error) {
	if err := c.EnsureValidToken(); err != nil {
		return nil, 0, err
	}

	url := c.BaseURL + endpoint
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Ocp-Apim-Subscription-Key", c.SubKey)
	req.Header.Set("Merchant-Serial-Number", c.MSN)

	// Set system information headers
	req.Header.Set("Vipps-System-Name", c.SystemName)
	req.Header.Set("Vipps-System-Version", c.SystemVersion)
	if c.SystemPluginName != "" {
		req.Header.Set("Vipps-System-Plugin-Name", c.SystemPluginName)
	}
	if c.SystemPluginVersion != "" {
		req.Header.Set("Vipps-System-Plugin-Version", c.SystemPluginVersion)
	}

	// Set idempotency key if provided
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var problemDetails struct {
			Title  string `json:"title"`
			Detail string `json:"detail"`
			Status int    `json:"status"`
			Code   string `json:"code"`
		}

		if err := json.Unmarshal(respBody, &problemDetails); err == nil {
			return respBody, resp.StatusCode, fmt.Errorf("API error: %s - %s (Code: %s, Status: %d)",
				problemDetails.Title, problemDetails.Detail, problemDetails.Code, problemDetails.Status)
		}

		return respBody, resp.StatusCode, fmt.Errorf("API error: status code %d, body: %s",
			resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}
