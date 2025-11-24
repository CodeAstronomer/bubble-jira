package types

type SelectLanguageData struct {
	ScreenHeight         int
	ScreenWidth          int
	Keys                 KeyData
	Strings              map[string]string
	AllowedLanguages     []string
	ChooseLanguageCursor any
}