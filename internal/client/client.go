package client

import (
	"context"
	"net/http"

	"github.com/ujjwalgoyal19/ado-dash/internal/auth"
)

// Client is the Azure DevOps REST client.
// baseURL is https://dev.azure.com/{org}
// releaseBaseURL is https://vsrm.dev.azure.com/{org} (Classic Release API)
type Client struct {
	baseURL        string
	releaseBaseURL string
	auth           auth.Provider
	http           *http.Client
}

// New returns a Client for the given org URL and auth provider.
func New(orgURL string, provider auth.Provider) *Client {
	return &Client{
		baseURL:        orgURL,
		releaseBaseURL: toReleaseURL(orgURL),
		auth:           provider,
		http:           &http.Client{},
	}
}

// toReleaseURL derives the vsrm base URL from the dev.azure.com org URL.
func toReleaseURL(orgURL string) string {
	// https://dev.azure.com/my-org → https://vsrm.dev.azure.com/my-org
	// Handled properly in implementation slice; placeholder for now.
	_ = orgURL
	return ""
}

// do executes an authenticated HTTP request.
func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	token, err := c.auth.Token(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return c.http.Do(req)
}
