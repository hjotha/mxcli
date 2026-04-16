// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client is a typed wrapper around the marketplace REST API. Callers
// obtain an authenticated http.Client via internal/auth.ClientFor and
// pass it here.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// New returns a marketplace client bound to the given HTTP client.
// The http.Client is expected to inject Mendix auth headers — use
// auth.ClientFor(ctx, profile) in production.
func New(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    BaseURL,
	}
}

// NewWithBaseURL constructs a client pointed at a specific base URL.
// Used by tests to redirect at httptest.Server.
func NewWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	return &Client{httpClient: httpClient, baseURL: baseURL}
}

// Search lists marketplace content matching a query. limit is the
// maximum number of results to return; pass 0 for the API default.
func (c *Client) Search(ctx context.Context, query string, limit int) (*ContentList, error) {
	q := url.Values{}
	if query != "" {
		q.Set("search", query)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	path := "/v1/content"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out ContentList
	if err := c.get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns the full detail for a single content item by ID.
func (c *Client) Get(ctx context.Context, contentID int) (*Content, error) {
	var out Content
	if err := c.get(ctx, fmt.Sprintf("/v1/content/%d", contentID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Versions returns all published versions for a content item, ordered
// newest first (per the API).
func (c *Client) Versions(ctx context.Context, contentID int) (*VersionList, error) {
	var out VersionList
	if err := c.get(ctx, fmt.Sprintf("/v1/content/%d/versions", contentID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) get(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("marketplace %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("marketplace %s: HTTP %d: %s", path, resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("marketplace %s: decode: %w", path, err)
	}
	return nil
}
