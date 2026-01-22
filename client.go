package flowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultAPIVersion = "v1.1"
	defaultTimeout    = 30 * time.Second
)

// Client represents a Flow Production Tracking API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiVersion string

	// Authentication
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time

	// Script credentials for re-authentication
	scriptName string
	scriptKey  string
}

// Config holds the configuration for creating a new Client
type Config struct {
	// SiteURL is your Flow site URL (e.g., "https://yoursite.shotgunstudio.com")
	SiteURL string

	// ScriptName is your API script name (client_id)
	ScriptName string

	// ScriptKey is your API script key (client_secret)
	ScriptKey string

	// APIVersion defaults to "v1.1" if not specified
	APIVersion string

	// HTTPClient allows you to provide a custom HTTP client
	HTTPClient *http.Client
}

// NewClient creates a new Flow API client
func NewClient(config Config) (*Client, error) {
	if config.SiteURL == "" {
		return nil, fmt.Errorf("site URL is required")
	}
	if config.ScriptName == "" {
		return nil, fmt.Errorf("script name is required")
	}
	if config.ScriptKey == "" {
		return nil, fmt.Errorf("script key is required")
	}

	apiVersion := config.APIVersion
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	client := &Client{
		baseURL:    config.SiteURL,
		apiVersion: apiVersion,
		httpClient: httpClient,
		scriptName: config.ScriptName,
		scriptKey:  config.ScriptKey,
	}

	// Authenticate immediately
	if err := client.authenticate(); err != nil {
		return nil, fmt.Errorf("initial authentication failed: %w", err)
	}

	return client, nil
}

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// authenticate obtains an access token using client credentials
func (c *Client) authenticate() error {
	authURL := fmt.Sprintf("%s/api/%s/auth/access_token", c.baseURL, c.apiVersion)

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.scriptName)
	data.Set("client_secret", c.scriptKey)

	req, err := http.NewRequest("POST", authURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.refreshToken = tokenResp.RefreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// GetAccessToken returns the current access token
// It will automatically re-authenticate if the token has expired
func (c *Client) GetAccessToken() (string, error) {
	// Check if token is expired or about to expire (with 60 second buffer)
	if time.Now().Add(60 * time.Second).After(c.tokenExpiry) {
		if err := c.authenticate(); err != nil {
			return "", err
		}
	}

	return c.accessToken, nil
}

// IsAuthenticated checks if the client has a valid access token
func (c *Client) IsAuthenticated() bool {
	return c.accessToken != "" && time.Now().Before(c.tokenExpiry)
}
