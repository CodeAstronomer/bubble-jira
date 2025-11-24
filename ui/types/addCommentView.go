package types

type AddCommentViewData struct {
	ScreenHeight        int
	ScreenWidth         int
	Keys                KeyData
	Strings             map[string]string
	AddCommentInput     bool
	AddCommentInputView string
	FocusedButton       string
	BlurredButton       string
}