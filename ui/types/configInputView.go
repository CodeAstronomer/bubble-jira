package types

type ConfigInputViewData struct {
	ScreenHeight         int
	ScreenWidth          int
	Keys                 KeyData
	Strings              map[string]string
	ConfigInput          bool
	ConfigInputView      string
	ConfigInputKey       string
	FocusedButton        string
	BlurredButton        string
}