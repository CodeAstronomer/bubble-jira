package ui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"bubble-jira/config"
)

func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
		height = 24
	}
	return width, height
}

// AppVersion stays constant
const AppVersion = "1.0.0"

var allowedLanguages = []string{"de-DE", "en-EN"}

// Strings holds all translated strings
var Strings map[string]string

// JiraStatusMap maps status strings to IDs
var JiraStatusMap map[string]int

// Initialize everything that depends on terminal size
var (
	cfg, _              = config.Load()
	Lang                = cfg.Lang
	width, height       = getTerminalSize()
	topBottomPadding    = 1
	leftRightPadding    = 2
	terminalWidth       = width - (leftRightPadding * 2)
	terminalHeight      = height - (topBottomPadding * 2)
	closeAfterSec       = 2
	closeAfterSecGit    = 0
	taskViewHeight      = 30
	autoFetchTimeSec    = 3
	keyLeft             = strings.TrimSpace(cfg.IssueKeyLoc)
	keyWrapper          = strings.TrimSpace(cfg.IssueKeyExample)
	exitKeys, keyExitKeysStr, keyUp, keyDown, keyFastUp, keyFastDown, keyEnter, keyNewComment = cfg.GetKeys()
	keyMap              = map[string]string{
		keyUp:       "↑",
		keyDown:     "↓",
		keyFastDown: "pgDown",
		keyFastUp:   "pgUp",
		keyEnter:    "⏎",
	}
    keyLeftRightMap      = map[string]string{
        "1": "left",
        "2": "right",
    }
    keyWrapperMap        = map[string]string{
        "1": "[XXX-XXXX]",
        "2": "(XXX-XXXX)",
    }
)

// loadStrings loads the JSON translation file for the current language
func loadStrings(lang string) {
	path := filepath.Join("languages", lang+".json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read language file '%s': %v", path, err))
	}
	err = json.Unmarshal(data, &Strings)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse language JSON: %v", err))
	}
}

// Init must be called at program start to load strings and setup maps
func Init() {
    if cfg == nil {
        panic("Config konnte nicht geladen werden")
    }

	loadStrings(Lang)

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
}
