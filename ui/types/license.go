package types

type LicenseViewData struct {
	Content      string
	IsLoading    bool
	Offset       int
	ScreenHeight int
	ScreenWidth  int
	Keys         KeyData
	Strings      map[string]string
}