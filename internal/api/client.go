package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const baseURL = "https://api.twitter.com/2"

// Client is the X API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new X API client
func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// NewClientWithBearerToken creates a client with bearer token auth
func NewClientWithBearerToken(bearerToken string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &bearerTransport{
				token: bearerToken,
				base:  http.DefaultTransport,
			},
		},
		baseURL: baseURL,
	}
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

// RateLimitError represents a rate limit response
type RateLimitError struct {
	RetryAfter int
	Message    string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// doRequest performs an HTTP request and decodes the response
func (c *Client) doRequest(ctx context.Context, method, path string, params url.Values, result interface{}) error {
	reqURL := c.baseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle rate limiting
	if resp.StatusCode == 429 {
		retryAfter := 60 // default 60 seconds
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if v, err := strconv.Atoi(ra); err == nil {
				retryAfter = v
			}
		}
		return &RateLimitError{
			RetryAfter: retryAfter,
			Message:    fmt.Sprintf("Rate limited. Try again in %d seconds", retryAfter),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// GetHomeTimeline fetches the authenticated user's home timeline
func (c *Client) GetHomeTimeline(ctx context.Context, maxResults int, paginationToken string) (*TimelineResponse, error) {
	// First get the authenticated user's ID
	me, err := c.GetMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	params := url.Values{}
	params.Set("tweet.fields", "created_at,public_metrics,author_id")
	params.Set("user.fields", "name,username,profile_image_url,verified,public_metrics")
	params.Set("expansions", "author_id")
	if maxResults > 0 {
		params.Set("max_results", fmt.Sprintf("%d", maxResults))
	}
	if paginationToken != "" {
		params.Set("pagination_token", paginationToken)
	}

	var result TimelineResponse
	path := fmt.Sprintf("/users/%s/timelines/reverse_chronological", me.ID)
	if err := c.doRequest(ctx, "GET", path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetMe returns the authenticated user
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	params := url.Values{}
	params.Set("user.fields", "name,username,description,profile_image_url,verified,public_metrics")

	var result struct {
		Data User `json:"data"`
	}
	if err := c.doRequest(ctx, "GET", "/users/me", params, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// GetUser fetches a user by username
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	params := url.Values{}
	params.Set("user.fields", "name,username,description,profile_image_url,verified,public_metrics")

	var result struct {
		Data User `json:"data"`
	}
	path := fmt.Sprintf("/users/by/username/%s", username)
	if err := c.doRequest(ctx, "GET", path, params, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// GetUserTweets fetches tweets from a user
func (c *Client) GetUserTweets(ctx context.Context, userID string, maxResults int, paginationToken string) (*TimelineResponse, error) {
	params := url.Values{}
	params.Set("tweet.fields", "created_at,public_metrics,author_id")
	params.Set("user.fields", "name,username,profile_image_url,verified")
	params.Set("expansions", "author_id")
	if maxResults > 0 {
		params.Set("max_results", fmt.Sprintf("%d", maxResults))
	}
	if paginationToken != "" {
		params.Set("pagination_token", paginationToken)
	}

	var result TimelineResponse
	path := fmt.Sprintf("/users/%s/tweets", userID)
	if err := c.doRequest(ctx, "GET", path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchTweets searches for tweets
func (c *Client) SearchTweets(ctx context.Context, query string, maxResults int, nextToken string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("tweet.fields", "created_at,public_metrics,author_id")
	params.Set("user.fields", "name,username,profile_image_url,verified")
	params.Set("expansions", "author_id")
	if maxResults > 0 {
		params.Set("max_results", fmt.Sprintf("%d", maxResults))
	}
	if nextToken != "" {
		params.Set("next_token", nextToken)
	}

	var result SearchResponse
	if err := c.doRequest(ctx, "GET", "/tweets/search/recent", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTweet fetches a single tweet by ID
func (c *Client) GetTweet(ctx context.Context, tweetID string) (*Tweet, *User, error) {
	params := url.Values{}
	params.Set("tweet.fields", "created_at,public_metrics,author_id")
	params.Set("user.fields", "name,username,profile_image_url,verified,public_metrics")
	params.Set("expansions", "author_id")

	var result struct {
		Data     Tweet    `json:"data"`
		Includes Includes `json:"includes"`
	}
	path := fmt.Sprintf("/tweets/%s", tweetID)
	if err := c.doRequest(ctx, "GET", path, params, &result); err != nil {
		return nil, nil, err
	}

	var author *User
	if len(result.Includes.Users) > 0 {
		author = &result.Includes.Users[0]
	}

	return &result.Data, author, nil
}
