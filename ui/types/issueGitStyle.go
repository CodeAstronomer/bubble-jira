package types

type IssueGitStyleData struct {
	ScreenHeight         int
	ScreenWidth          int
	Keys                 KeyData
	Strings              map[string]string
	Style1               string
	Style2               string
	GitIssueStyleCursor  any
}