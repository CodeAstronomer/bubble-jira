
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

// SearchResponse represents the full Jira search response with pagination.
type SearchResponse struct {
	Issues        []IssueData `json:"issues"`
	NextPageToken string      `json:"nextPageToken"`
	IsLast        bool        `json:"isLast"`
}

// IssueData represents a single issue in the search response.
type IssueData struct {
	Expand string `json:"expand"`
	ID     string `json:"id"`
	Self   string `json:"self"`
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

// Author represents the comment author
type Author struct {
	AccountID    string `json:"accountId"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
	Active       bool   `json:"active"`
	TimeZone     string `json:"timeZone"`
}

// Comment represents a Jira comment with replies
type Comment struct {
	ID            string     `json:"id"`
	Self          string     `json:"self"`
	Author        Author     `json:"author"`
	Body          interface{} `json:"body"`
	BodyJSON      string     `json:"-"`
	Created       string     `json:"created"`
	Updated       string     `json:"updated"`
	UpdateAuthor  Author     `json:"updateAuthor"`
	JSDPublic     bool       `json:"jsdPublic"`
	Replies       []Comment  `json:"-"`
}

// CommentsResponse represents the Jira comments API response
type CommentsResponse struct {
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	Comments   []Comment `json:"comments"`
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
	u.Path = strings.TrimRight(u.Path, "/") + "/rest/api/3/search/jql"
	q := url.Values{}
	q.Set("jql", c.cfg.JQL)
	q.Set("maxResults", "50")
	q.Set("fields", "summary,status,id,self,key")
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
	var response SearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Map to our Issue type
	issues := make([]Issue, len(response.Issues))
	for i, is := range response.Issues {
		issues[i] = Issue{
			Key:     is.Key,
			Title:   is.Fields.Summary,
			Status:  is.Fields.Status.Name,
			HTMLURL: fmt.Sprintf("%s/browse/%s", strings.TrimRight(c.cfg.BaseURL, "/"), is.Key),
		}
	}

	return issues, nil
}

// FetchComments fetches all comments for a specific issue
func (c *Client) FetchComments(ctx context.Context, issueKey string) ([]Comment, error) {
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

	// Build issue URL to get comments
	u.Path = strings.TrimRight(u.Path, "/") + "/rest/api/3/issue/" + issueKey
	q := url.Values{}
	q.Set("fields", "comment")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.cfg.Email, c.cfg.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira API returned status %s", resp.Status)
	}

	// Parse the response to extract comments
	var issueResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issueResponse); err != nil {
		return nil, err
	}

	fields, ok := issueResponse["fields"].(map[string]interface{})
	if !ok {
		return []Comment{}, nil
	}

	commentData, ok := fields["comment"].(map[string]interface{})
	if !ok {
		return []Comment{}, nil
	}

	commentsArray, ok := commentData["comments"].([]interface{})
	if !ok {
		return []Comment{}, nil
	}

	comments := make([]Comment, 0)
	for _, commentInterface := range commentsArray {
		commentMap, ok := commentInterface.(map[string]interface{})
		if !ok {
			continue
		}

		comment, err := parseCommentFromMap(commentMap)
		if err == nil {
			// Convert body to JSON string
			if comment.Body != nil {
				bodyBytes, _ := json.Marshal(comment.Body)
				comment.BodyJSON = string(bodyBytes)
			}
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

// parseCommentFromMap converts a map to a Comment struct
func parseCommentFromMap(commentMap map[string]interface{}) (Comment, error) {
	comment := Comment{}

	// Parse basic fields
	if id, ok := commentMap["id"].(string); ok {
		comment.ID = id
	}

	if self, ok := commentMap["self"].(string); ok {
		comment.Self = self
	}

	if created, ok := commentMap["created"].(string); ok {
		comment.Created = created
	}

	if updated, ok := commentMap["updated"].(string); ok {
		comment.Updated = updated
	}

	if jsdPublic, ok := commentMap["jsdPublic"].(bool); ok {
		comment.JSDPublic = jsdPublic
	}

	// Parse body (keep as interface for JSON marshaling)
	if body, ok := commentMap["body"]; ok {
		comment.Body = body
	}

	// Parse author
	if authorData, ok := commentMap["author"].(map[string]interface{}); ok {
		comment.Author = parseAuthor(authorData)
	}

	// Parse update author
	if updateAuthorData, ok := commentMap["updateAuthor"].(map[string]interface{}); ok {
		comment.UpdateAuthor = parseAuthor(updateAuthorData)
	}

	return comment, nil
}

// parseAuthor extracts author information from a map
func parseAuthor(authorMap map[string]interface{}) Author {
	author := Author{}

	if accountID, ok := authorMap["accountId"].(string); ok {
		author.AccountID = accountID
	}
	if emailAddress, ok := authorMap["emailAddress"].(string); ok {
		author.EmailAddress = emailAddress
	}
	if displayName, ok := authorMap["displayName"].(string); ok {
		author.DisplayName = displayName
	}
	if active, ok := authorMap["active"].(bool); ok {
		author.Active = active
	}
	if timeZone, ok := authorMap["timeZone"].(string); ok {
		author.TimeZone = timeZone
	}

	return author
}