package types

type CommentViewData struct {
	IsLoading           bool
	ScreenHeight        int
	ScreenWidth         int
	Keys                KeyData
	Strings             map[string]string
	CommentsViewPort    string
}