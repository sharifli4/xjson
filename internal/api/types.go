package api

import "time"

// Tweet represents a tweet from the X API
type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	Metrics   *Metrics  `json:"public_metrics,omitempty"`
}

// Metrics represents tweet engagement metrics
type Metrics struct {
	RetweetCount int `json:"retweet_count"`
	ReplyCount   int `json:"reply_count"`
	LikeCount    int `json:"like_count"`
	QuoteCount   int `json:"quote_count"`
	Impressions  int `json:"impression_count"`
}

// User represents an X user
type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	Description     string `json:"description,omitempty"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
	Verified        bool   `json:"verified,omitempty"`
	FollowersCount  int    `json:"followers_count,omitempty"`
	FollowingCount  int    `json:"following_count,omitempty"`
	TweetCount      int    `json:"tweet_count,omitempty"`
}

// TimelineResponse represents the API response for timeline
type TimelineResponse struct {
	Data     []Tweet         `json:"data"`
	Includes *Includes       `json:"includes,omitempty"`
	Meta     *ResponseMeta   `json:"meta,omitempty"`
}

// Includes contains expanded objects
type Includes struct {
	Users []User `json:"users,omitempty"`
}

// ResponseMeta contains pagination info
type ResponseMeta struct {
	ResultCount   int    `json:"result_count"`
	NextToken     string `json:"next_token,omitempty"`
	PreviousToken string `json:"previous_token,omitempty"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Data     []Tweet       `json:"data"`
	Includes *Includes     `json:"includes,omitempty"`
	Meta     *ResponseMeta `json:"meta,omitempty"`
}
