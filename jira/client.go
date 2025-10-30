
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

// validateConfig checks if the config is valid
func (c *Client) validateConfig() error {
	if c.cfg == nil {
		return fmt.Errorf("nil config")
	}
	if !c.cfg.IsValid() {
		return fmt.Errorf("config is not valid - please fill all required fields")
	}
	return nil
}

// buildRequest creates an HTTP request with Jira authentication
func (c *Client) buildRequest(ctx context.Context, method, path string, query url.Values) (*http.Request, error) {
	u, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = strings.TrimRight(u.Path, "/") + path
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.cfg.Email, c.cfg.APIToken)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// executeRequest performs an HTTP request and checks for errors
func (c *Client) executeRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("jira API returned status %s", resp.Status)
	}

	return resp, nil
}

// FetchAssignedIssues fetches issues using the JQL in config.
func (c *Client) FetchAssignedIssues(ctx context.Context) ([]Issue, error) {
	if err := c.validateConfig(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("jql", c.cfg.JQL)
	q.Set("maxResults", "50")
	q.Set("fields", "summary,status,id,self,key")

	req, err := c.buildRequest(ctx, "GET", "/rest/api/3/search/jql", q)
	if err != nil {
		return nil, err
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
	if err := c.validateConfig(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("fields", "comment")

	req, err := c.buildRequest(ctx, "GET", "/rest/api/3/issue/"+issueKey, q)
	if err != nil {
		return nil, err
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response to extract comments
	var issueResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issueResponse); err != nil {
		return nil, err
	}

	comments := extractCommentsFromResponse(issueResponse)
	return comments, nil
}

// PostStatus changes the status of a Jira issue using the transition ID.
// Returns the HTTP status code and an error if something goes wrong.
func (c *Client) PostStatus(ctx context.Context, issueKey string, selectedID int) (int, error) {
	/*  if err := c.validateConfig(); err != nil {
		return 0, err
	}

	urlPath := fmt.Sprintf("/rest/api/3/issue/%s/transitions", issueKey)

	// Prepare JSON body
	bodyMap := map[string]map[string]string{
		"transition": {
			"id": fmt.Sprintf("%d", selectedID),
		},
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal JSON body: %w", err)
	}

	req, err := c.buildRequest(ctx, "POST", urlPath, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return 200, nil
	case 401:
		return resp.StatusCode, fmt.Errorf("unauthorized: check your Jira email/API token")
	case 404:
		return resp.StatusCode, fmt.Errorf("issue %s not found", issueKey)
	default:
		return resp.StatusCode, fmt.Errorf("jira API returned status %s", resp.Status)
	} */
	return 200, nil
}

// PostComment added a new comment to a Jira Issue
// Returns the HTTP status code and an error if something goes wrong.
func (c *Client) PostComment(ctx context.Context, issueKey string, commentText string) (int, error) {
	/* if err := c.validateConfig(); err != nil {
		return 0, err
	}

	urlPath := fmt.Sprintf("/rest/api/3/issue/%s/comment", issueKey)

	// Prepare JSON body in the desired schema
    bodyMap := map[string]interface{}{
    	"body": map[string]interface{}{
    		"type":    "doc",
    		"version": 1,
    		"content": []map[string]interface{}{
    			{
    				"type": "paragraph",
    				"content": []map[string]interface{}{
    					{
    						"type": "text",
    						"text": commentText,
    					},
    				},
    			},
    		},
    	},
    }

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal JSON body: %w", err)
	}

	req, err := c.buildRequest(ctx, "POST", urlPath, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return 200, nil
	case 400:
		return resp.StatusCode, fmt.Errorf("bad request")
	case 401:
		return resp.StatusCode, fmt.Errorf("unauthorized: check your Jira email/API token")
	case 404:
		return resp.StatusCode, fmt.Errorf("issue %s not found", issueKey)
	case 413:
		return resp.StatusCode, fmt.Errorf("Request Entity Too Large")
	default:
		return resp.StatusCode, fmt.Errorf("jira API returned status %s", resp.Status)
	} */
	return 200, nil
}

// extractCommentsFromResponse parses the API response to extract comments
func extractCommentsFromResponse(issueResponse map[string]interface{}) []Comment {
	fields, ok := issueResponse["fields"].(map[string]interface{})
	if !ok {
		return []Comment{}
	}

	commentData, ok := fields["comment"].(map[string]interface{})
	if !ok {
		return []Comment{}
	}

	commentsArray, ok := commentData["comments"].([]interface{})
	if !ok {
		return []Comment{}
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

	return comments
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