// Package github handles GitHub API integration.
// Phase 5 implementation.
package github

// Client handles GitHub API operations.
type Client struct {
	token string
}

// NewClient creates a new GitHub client.
func NewClient(token string) *Client {
	return &Client{token: token}
}
