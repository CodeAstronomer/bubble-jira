package types

type EnterCommitMessageData struct {
	ScreenHeight         int
	ScreenWidth          int
	Keys                 KeyData
	Strings              map[string]string
	CommitGitMessage     bool
	CommitGitMessageView string
	FocusedButtonGit     string
	BlurredButtonGit     string
}