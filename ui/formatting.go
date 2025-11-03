package ui

import (
	"bubble-jira/jira"
	"encoding/json"
	"strings"
	"time"
)

// ContentNode represents a node in Jira's structured content format
type ContentNode struct {
	Type    string                 `json:"type"`
	Text    string                 `json:"text,omitempty"`
	Content []ContentNode          `json:"content,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
}

// CommentBody represents the body structure of a Jira comment
type CommentBody struct {
	Content []ContentNode `json:"content"`
}

// truncateString truncates a string to maxLen characters and adds ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatTime formats a time string from Jira format to a readable format
func formatTime(timeStr string) string {
	// Parse the time string from Jira (e.g., "2025-10-28T09:31:21.319+0100")
	t, err := time.Parse("2006-01-02T15:04:05.000-0700", timeStr)
	if err != nil {
		// Try alternative format
		t, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return timeStr
		}
	}
	return t.Format("2006-01-02 15:04 MST")
}

// parseContentNodes recursively parses content nodes and builds a string
func parseContentNodes(nodes []ContentNode, result *strings.Builder, depth int) {
	for _, node := range nodes {
		switch node.Type {
		case "paragraph":
			for _, child := range node.Content {
				switch child.Type {
				case "text":
					result.WriteString(child.Text)
				case "mention":
					if text, ok := child.Attrs["text"].(string); ok {
						result.WriteString(text)
					}
				case "hardBreak":
					result.WriteString("\n")
				}
			}
			result.WriteString("\n")

		case "bulletList", "orderedList":
			for _, child := range node.Content {
				if child.Type == "listItem" {
					result.WriteString("• ")
					parseContentNodes(child.Content, result, depth+1)
				}
			}

		case "codeBlock":
			result.WriteString("```\n")
			for _, child := range node.Content {
				if child.Type == "text" {
					result.WriteString(child.Text)
				}
			}
			result.WriteString("\n```\n")

		case "heading":
			if level, ok := node.Attrs["level"].(float64); ok {
				for i := 0; i < int(level); i++ {
					result.WriteString("#")
				}
				result.WriteString(" ")
			}
			for _, child := range node.Content {
				if child.Type == "text" {
					result.WriteString(child.Text)
				}
			}
			result.WriteString("\n")

		case "text":
			result.WriteString(node.Text)

		default:
			// Recursively parse unknown types
			if len(node.Content) > 0 {
				parseContentNodes(node.Content, result, depth)
			}
		}
	}
}

// parseCommentBody parses the body JSON from a Jira comment
func parseCommentBody(bodyJSON string) string {
	var commentBody CommentBody
	if err := json.Unmarshal([]byte(bodyJSON), &commentBody); err != nil {
		// If it's not valid JSON, return as-is
		return bodyJSON
	}

	var result strings.Builder
	parseContentNodes(commentBody.Content, &result, 0)

	return strings.TrimSpace(result.String())
}

// renderCommentsToMarkdown converts comments to markdown format
func renderCommentsToMarkdown(comments []jira.Comment) string {
	if len(comments) == 0 {
		return Strings["noComments"]
	}

	var result strings.Builder
	result.Grow(len(comments) * 256) // Pre-allocate based on comment count

	for _, comment := range comments {
		authorName := comment.Author.DisplayName
		if authorName == "" {
			authorName = comment.Author.EmailAddress
		}

		result.WriteString("**")
		result.WriteString(authorName)
		result.WriteString("** - ")
		result.WriteString(formatTime(comment.Created))
		result.WriteString("\n\n")

		// Convert body to JSON string for parsing
		var bodyStr string
		if comment.BodyJSON != "" {
			bodyStr = comment.BodyJSON
		} else if comment.Body != nil {
			bodyBytes, _ := json.Marshal(comment.Body)
			bodyStr = string(bodyBytes)
		}

		if bodyStr != "" {
			result.WriteString(parseCommentBody(bodyStr))
			result.WriteString("\n\n")
		}

		result.WriteString("---\n\n")
	}

	return result.String()
}