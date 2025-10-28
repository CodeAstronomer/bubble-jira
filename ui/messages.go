
package ui

import (
	"time"

	"bubble-jira/jira"
)

// issuesFetchedMsg message for when issues are fetched
type issuesFetchedMsg struct {
	issues []jira.Issue
	err    error
}

// commentsFetchedMsg message for when comments are fetched
type commentsFetchedMsg struct {
	comments []jira.Comment
	err      error
}

// licenceLoadedMsg message for when license is loaded
type licenceLoadedMsg struct {
	content string
	err     error
}

// tickMsg message for progress updates
type tickMsg time.Time