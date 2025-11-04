package ui

import (
    "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
	"bubble-jira/config"
	"regexp"
)
//go:embed languages
var languageFS embed.FS

func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
		height = 24
	}
	return width, height
}

// AppVersion stays constant
const AppVersion = "1.0.1"

// Initialize everything that depends on terminal size
var (
	cfg                   *config.Config
	Lang                  string
	width, height       = getTerminalSize()
	topBottomPadding    = 1
	leftRightPadding    = 2
	terminalWidth       = width - (leftRightPadding * 2)
	terminalHeight      = height - (topBottomPadding * 2)
	closeAfterSec       = 2
	closeAfterSecGit    = 0
	taskViewHeight      = 30
	autoFetchTimeSec    = 3
	keyLeft               string
	keyWrapper            string
	style1              = "[XXX-XXXX]"
	style2              = "(XXX-XXXX)"
	allowedLanguages    = []string{"de-DE", "en-EN"}
    languageRegex       = regexp.MustCompile(`^[a-z]{2}-[A-Z]{2}$`)
    statuses            []string

	exitKeys        []string
    keyExitKeysStr  string
    keyUp           string
    keyDown         string
    keyFastUp       string
    keyFastDown     string
    keyEnter        string
    keyNewComment   string
    keySearch       string
    keyForceQuit    string
    Strings map[string]string
	keyMap map[string]string
    keyLeftRightMap map[string]string
    keyWrapperMap map[string]string
    JiraStatusMap map[string]int
)

// loadStrings loads the JSON translation file for the current language
func loadStrings(lang string) {
	path := "languages/" + lang + ".json"
	data, err := languageFS.ReadFile(path)
	if err != nil {
        panic(fmt.Sprintf("Failed to read embedded language file '%s': %v", path, err))
    }
    err = json.Unmarshal(data, &Strings)
    if err != nil {
        panic(fmt.Sprintf("Failed to parse language JSON: %v", err))
    }
}

// Init must be called at program start to load strings and setup maps
func Init() {
    var err error
    cfg, err = config.Load()
    if err != nil {
        panic(fmt.Sprintf("Failed to load config: %v", err))
    }

    Lang = cfg.Lang
    keyLeft = strings.TrimSpace(cfg.IssueKeyLoc)
    keyWrapper = strings.TrimSpace(cfg.IssueKeyStyle)
    exitKeys, keyExitKeysStr, keyUp, keyDown, keyFastUp, keyFastDown, keyEnter, keyNewComment, keySearch, keyForceQuit = cfg.GetKeys()

	loadStrings(Lang)

	statuses = []string{
        Strings["Open"],
        Strings["CurrentlyInProgress"],
        Strings["Done"],
        Strings["Reopened"],
        Strings["Closed"],
        Strings["Backlog"],
        Strings["QM"],
        Strings["Waiting"],
        Strings["Staging"],
    }

	// Initialize Jira status map dynamically using Strings
	JiraStatusMap = map[string]int{
		Strings["Open"]:                11,
		Strings["CurrentlyInProgress"]: 21,
		Strings["Done"]:                31,
		Strings["Reopened"]:            41,
		Strings["Closed"]:              51,
		Strings["Backlog"]:             61,
		Strings["QM"]:                  71,
		Strings["Waiting"]:             81,
		Strings["Staging"]:             91,
	}

    keyMap = map[string]string{
        keyUp:       "↑",
        keyDown:     "↓",
        keyFastDown: "pgDown",
        keyFastUp:   "pgUp",
        keyEnter:    "⏎",
    }

    keyLeftRightMap = map[string]string{
        "1": Strings["left"],
        "2": Strings["right"],
    }

    keyWrapperMap = map[string]string{
        "1": style1,
        "2": style2,
    }
}
