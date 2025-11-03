package ui

import (
    _ "embed"
    "context"
    "time"
    "fmt"

	"bubble-jira/jira"

	tea "github.com/charmbracelet/bubbletea"
)
//go:embed LICENSE
var licenseContent []byte

type statusCodeMsg struct {
	Code int
	Err  error
}

// fetchJiraTasksCmd creates a command to fetch Jira tasks
func fetchJiraTasksCmd(jc *jira.Client) tea.Cmd {
	return func() tea.Msg {
		issues, err := jc.FetchAssignedIssues(context.Background())
		return issuesFetchedMsg{issues: issues, err: err}
	}
}

// fetchCommentsCmd creates a command to fetch comments for an issue
func fetchCommentsCmd(jc *jira.Client, issueKey string) tea.Cmd {
    return func() tea.Msg {
        comments, err := jc.FetchComments(context.Background(), issueKey)
        return commentsFetchedMsg{
            taskKey:  issueKey,
            comments: comments,
            err:      err,
        }
    }
}

// UpdateState of Jira Task
func fetchStatusCmd(jc *jira.Client, issueKey string, selectedID int) tea.Cmd {
	return func() tea.Msg {
		statusCode, _ := jc.PostStatus(context.Background(), issueKey, selectedID)
		return statusCodeMsg{Code: statusCode}
	}
}

// createNewComment
func createNewComment(jc *jira.Client, issueKey string, commentText string) tea.Cmd {
	return func() tea.Msg {
		statusCode, _ := jc.PostComment(context.Background(), issueKey, commentText)
		return statusCodeMsg{Code: statusCode}
	}
}

// tickFetchCmd creates a command for progress updates
func tickFetchCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// loadLicenceCmd creates a command to load the license
func loadLicenceCmd() tea.Cmd {
	return func() tea.Msg {
	    content, err := loadLicenceContent()
        return licenceLoadedMsg{content: content, err: err}
	}
}

// loadLicenceContent loads the license content from various sources
func loadLicenceContent() (string, error) {
	if len(licenseContent) == 0 {
        return "Embedded LICENSE file not found or is empty.", fmt.Errorf("embedded license is empty")
    }
    return string(licenseContent), nil
}

 // hoverTimeoutCmd erstellt einen Command für den Hover-Timer
func hoverTimeoutCmd(taskKey string, duration time.Duration) tea.Cmd {
    return func() tea.Msg {
        time.Sleep(duration)
        return hoverTimeoutMsg{taskKey: taskKey}
    }
}


// fetchCommentsBackgroundCmd fetched Comments im Hintergrund für Caching
func fetchCommentsBackgroundCmd(jc *jira.Client, taskKey string) tea.Cmd {
    return func() tea.Msg {
        comments, err := jc.FetchComments(context.Background(), taskKey)
        return commentsCachedMsg{
            taskKey:  taskKey,
            comments: comments,
            err:      err,
        }
    }
}