package types

type IssueGitLocationData struct {
	ScreenHeight         int
	ScreenWidth          int
	Keys                 KeyData
	Strings              map[string]string
	GitIssueLocCursor    any
}