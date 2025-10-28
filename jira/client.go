package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"bubble-jira/config"
)

// Issue is a compact representation used in the UI.
type Issue struct {
	Key     string
	Title   string
	Status  string
	HTMLURL string
}

// Client knows how to talk to Jira REST API.
type Client struct {
	cfg *config.Config
	hc  *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg, hc: http.DefaultClient}
}

// FetchAssignedIssues fetches issues using the JQL in config.
func (c *Client) FetchAssignedIssues(ctx context.Context) ([]Issue, error) {
	if c.cfg == nil {
		return nil, fmt.Errorf("nil config")
	}
	if !c.cfg.IsValid() {
		return nil, fmt.Errorf("config is not valid - please fill all required fields")
	}

	u, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	// Build search URL
	u.Path = strings.TrimRight(u.Path, "/") + "/rest/api/3/search"
	q := url.Values{}
	q.Set("jql", c.cfg.JQL)
	q.Set("maxResults", "200")
	q.Set("fields", "summary,status")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	// Basic auth with email:apiToken
	req.SetBasicAuth(c.cfg.Email, c.cfg.APIToken)
    req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira API returned status %s", resp.Status)
	}

	// Parse the response JSON
	var data struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Name string `json:"name"`
				} `json:"status"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Map to our Issue type
	issues := make([]Issue, len(data.Issues))
	for i, is := range data.Issues {
		issues[i] = Issue{
			Key:     is.Key,
			Title:   is.Fields.Summary,
			Status:  is.Fields.Status.Name,
			HTMLURL: fmt.Sprintf("%s/browse/%s", strings.TrimRight(c.cfg.BaseURL, "/"), is.Key),
		}
	}

	return issues, nil
}