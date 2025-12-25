package transform

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kenan/xjson/internal/api"
)

// DisguisedPayload represents a tweet in "API response" format
type DisguisedPayload struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Endpoint  string                 `json:"endpoint"`
	Status    int                    `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// DisguisedResponse wraps multiple tweets as API responses
type DisguisedResponse struct {
	Method     string             `json:"method"`
	Endpoint   string             `json:"endpoint"`
	StatusCode int                `json:"status_code"`
	Latency    string             `json:"latency_ms"`
	Data       []DisguisedPayload `json:"data"`
	Meta       *MetaInfo          `json:"_meta,omitempty"`
}

// MetaInfo contains pagination metadata
type MetaInfo struct {
	ResultCount int    `json:"result_count"`
	NextCursor  string `json:"next_cursor,omitempty"`
	HasMore     bool   `json:"has_more"`
}

// TransformTweet converts a tweet to disguised format
func TransformTweet(tweet *api.Tweet, author *api.User) DisguisedPayload {
	payload := map[string]interface{}{
		"content": tweet.Text,
		"author": map[string]interface{}{
			"handle":       author.Username,
			"display_name": author.Name,
			"verified":     author.Verified,
		},
		"created_at": tweet.CreatedAt.Format(time.RFC3339),
	}

	if tweet.Metrics != nil {
		payload["metrics"] = map[string]interface{}{
			"impressions": tweet.Metrics.Impressions,
			"engagements": tweet.Metrics.LikeCount + tweet.Metrics.RetweetCount + tweet.Metrics.ReplyCount,
			"retweets":    tweet.Metrics.RetweetCount,
			"likes":       tweet.Metrics.LikeCount,
			"replies":     tweet.Metrics.ReplyCount,
		}
	}

	return DisguisedPayload{
		ID:        tweet.ID,
		Type:      "status_update",
		Endpoint:  fmt.Sprintf("/v2/statuses/%s", tweet.ID),
		Status:    200,
		Timestamp: time.Now().Format(time.RFC3339),
		Payload:   payload,
	}
}

// TransformTimeline converts a timeline response to disguised format
func TransformTimeline(resp *api.TimelineResponse, endpoint string) *DisguisedResponse {
	userMap := make(map[string]*api.User)
	if resp.Includes != nil {
		for i := range resp.Includes.Users {
			userMap[resp.Includes.Users[i].ID] = &resp.Includes.Users[i]
		}
	}

	data := make([]DisguisedPayload, 0, len(resp.Data))
	for _, tweet := range resp.Data {
		author := userMap[tweet.AuthorID]
		if author == nil {
			author = &api.User{Username: "unknown", Name: "Unknown User"}
		}
		data = append(data, TransformTweet(&tweet, author))
	}

	result := &DisguisedResponse{
		Method:     "GET",
		Endpoint:   endpoint,
		StatusCode: 200,
		Latency:    fmt.Sprintf("%d", 50+len(resp.Data)*2),
		Data:       data,
	}

	if resp.Meta != nil {
		result.Meta = &MetaInfo{
			ResultCount: resp.Meta.ResultCount,
			NextCursor:  resp.Meta.NextToken,
			HasMore:     resp.Meta.NextToken != "",
		}
	}

	return result
}

// TransformUser converts a user to disguised format
func TransformUser(user *api.User) DisguisedPayload {
	payload := map[string]interface{}{
		"handle":       user.Username,
		"display_name": user.Name,
		"bio":          user.Description,
		"verified":     user.Verified,
		"avatar_url":   user.ProfileImageURL,
		"stats": map[string]interface{}{
			"followers":  user.FollowersCount,
			"following":  user.FollowingCount,
			"posts":      user.TweetCount,
		},
	}

	return DisguisedPayload{
		ID:        user.ID,
		Type:      "user_profile",
		Endpoint:  fmt.Sprintf("/v2/users/%s", user.ID),
		Status:    200,
		Timestamp: time.Now().Format(time.RFC3339),
		Payload:   payload,
	}
}

// TransformSearch converts search results to disguised format
func TransformSearch(resp *api.SearchResponse, query string) *DisguisedResponse {
	userMap := make(map[string]*api.User)
	if resp.Includes != nil {
		for i := range resp.Includes.Users {
			userMap[resp.Includes.Users[i].ID] = &resp.Includes.Users[i]
		}
	}

	data := make([]DisguisedPayload, 0, len(resp.Data))
	for _, tweet := range resp.Data {
		author := userMap[tweet.AuthorID]
		if author == nil {
			author = &api.User{Username: "unknown", Name: "Unknown User"}
		}
		data = append(data, TransformTweet(&tweet, author))
	}

	result := &DisguisedResponse{
		Method:     "GET",
		Endpoint:   fmt.Sprintf("/v2/search?q=%s", query),
		StatusCode: 200,
		Latency:    fmt.Sprintf("%d", 80+len(resp.Data)*3),
		Data:       data,
	}

	if resp.Meta != nil {
		result.Meta = &MetaInfo{
			ResultCount: resp.Meta.ResultCount,
			NextCursor:  resp.Meta.NextToken,
			HasMore:     resp.Meta.NextToken != "",
		}
	}

	return result
}

// ToJSON converts a payload to pretty-printed JSON
func ToJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToCompactJSON converts a payload to compact JSON
func ToCompactJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
