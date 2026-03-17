package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenResponse represents the JSON response from Keycloak token endpoint
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	IdToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

// KeycloakClient defines the interface for interacting with Keycloak
type KeycloakClient interface {
	Login(ctx context.Context, username, password string) (*TokenResponse, error)
}

// keycloakClientImpl implements KeycloakClient
type keycloakClientImpl struct {
	baseURL      string
	realm        string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewKeycloakClient creates a new Keycloak client
func NewKeycloakClient(baseURL, realm, clientID, clientSecret string) KeycloakClient {
	// Remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &keycloakClientImpl{
		baseURL:      baseURL,
		realm:        realm,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Login performs Resource Owner Password Credentials Grant to get a token
func (c *keycloakClientImpl) Login(ctx context.Context, username, password string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.baseURL, c.realm)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.clientID)
	if c.clientSecret != "" {
		data.Set("client_secret", c.clientSecret)
	}
	data.Set("username", username)
	data.Set("password", password)
	// Usually scope "openid" is required for OIDC
	data.Set("scope", "openid")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("keycloak returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("unmarshal token response failed: %w", err)
	}

	return &tokenResp, nil
}
