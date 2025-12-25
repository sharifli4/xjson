package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

const (
	authURL  = "https://twitter.com/i/oauth2/authorize"
	tokenURL = "https://api.twitter.com/2/oauth2/token"
)

// TokenStore handles OAuth token persistence
type TokenStore struct {
	configPath string
}

// StoredToken represents a persisted OAuth token
type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	home, _ := os.UserHomeDir()
	return &TokenStore{
		configPath: filepath.Join(home, ".xjson_token.json"),
	}
}

// Save persists the OAuth token
func (ts *TokenStore) Save(token *oauth2.Token) error {
	stored := StoredToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ts.configPath, data, 0600)
}

// Load retrieves the stored OAuth token
func (ts *TokenStore) Load() (*oauth2.Token, error) {
	data, err := os.ReadFile(ts.configPath)
	if err != nil {
		return nil, err
	}

	var stored StoredToken
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  stored.AccessToken,
		RefreshToken: stored.RefreshToken,
		TokenType:    stored.TokenType,
		Expiry:       stored.Expiry,
	}, nil
}

// Exists checks if a stored token exists
func (ts *TokenStore) Exists() bool {
	_, err := os.Stat(ts.configPath)
	return err == nil
}

// generateCodeVerifier creates a PKCE code verifier
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge creates a PKCE code challenge from verifier
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// Authenticator handles OAuth 2.0 PKCE flow
type Authenticator struct {
	config     *oauth2.Config
	tokenStore *TokenStore
}

// NewAuthenticator creates a new OAuth authenticator
func NewAuthenticator(clientID, clientSecret, redirectURL string) *Authenticator {
	return &Authenticator{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
			RedirectURL: redirectURL,
			Scopes:      []string{"tweet.read", "users.read", "offline.access"},
		},
		tokenStore: NewTokenStore(),
	}
}

// GetToken returns a valid token, refreshing if necessary
func (a *Authenticator) GetToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := a.tokenStore.Load()
	if err != nil {
		return nil, fmt.Errorf("no stored token: %w", err)
	}

	// Check if token needs refresh
	if token.Expiry.Before(time.Now()) && token.RefreshToken != "" {
		src := a.config.TokenSource(ctx, token)
		newToken, err := src.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		if err := a.tokenStore.Save(newToken); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %w", err)
		}
		return newToken, nil
	}

	return token, nil
}

// StartAuthFlow initiates the OAuth flow and returns the auth URL
func (a *Authenticator) StartAuthFlow() (authURL string, verifier string, err error) {
	verifier, err = generateCodeVerifier()
	if err != nil {
		return "", "", err
	}

	challenge := generateCodeChallenge(verifier)

	authURL = a.config.AuthCodeURL("state",
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return authURL, verifier, nil
}

// CompleteAuthFlow exchanges the auth code for a token
func (a *Authenticator) CompleteAuthFlow(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	token, err := a.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return nil, err
	}

	if err := a.tokenStore.Save(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

// HasStoredToken checks if there's a valid stored token
func (a *Authenticator) HasStoredToken() bool {
	return a.tokenStore.Exists()
}

// HTTPClient returns an HTTP client with the OAuth token
func (a *Authenticator) HTTPClient(ctx context.Context) (*http.Client, error) {
	token, err := a.GetToken(ctx)
	if err != nil {
		return nil, err
	}
	return a.config.Client(ctx, token), nil
}
