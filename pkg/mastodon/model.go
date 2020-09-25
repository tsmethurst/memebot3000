package mastodon

// TokenResponse represents a token response from /oauth/token
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	CreatedAt   int64  `json:"created_at"`
}

// MediaResponse represents a media upload response from /api/v2/media
type MediaResponse struct {
	MediaID string `json:"id"`
}
