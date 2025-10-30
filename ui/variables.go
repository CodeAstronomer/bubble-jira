package ui

import (
    "os"
	"golang.org/x/term"

	"bubble-jira/config"
)

func getTerminalSize() (int, int) {
    width, height, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        width = 80
    }
    return width, height
}

const (
    AppVersion = "1.0.0"

	MenuViewTasksTitle  = "View Tasks"
	MenuSettingsTitle   = "Settings"
	MenuQuitTitle       = "Quit"
	MenuConfigTitle     = "Edit Config"
	MenuLicenceTitle    = "View Licence"

	MenuBackTitle       = "Back"
	ContextViewComments = "View Comments"

	Lang                = "de"

	Open                = "Offen"
	CurrentlyInProgress = "Gerade in Arbeit"
	Done                = "Erledigt"
	Reopened            = "Wiedereröffnet"
	Closed              = "Geschlossen"
	Backlog             = "Backlog"
	QM                  = "QM"
	Waiting             = "Warten auf ..."
	Staging             = "Zur Abnahme im Staging"
)

var (
    // General
    width, height      = getTerminalSize()
    topBottomPadding   = 1
    leftRightPadding   = 2
    terminalWidth      = (width-(leftRightPadding*leftRightPadding))
    terminalHeight     = (height-(topBottomPadding*topBottomPadding))
    closeAfterSec      = 2

    //Tasks
    taskViewHeight = 30
    autoFetchTimeSec = 3

    //Jira
    jiraStatusMap = map[string]int{
        Open: 11,
        CurrentlyInProgress: 21,
        Done: 31,
        Reopened: 41,
        Closed: 51,
        Backlog: 61,
        QM: 71,
        Waiting: 81,
        Staging: 91,
    }

    //Help Footer
    cfg = config.DefaultConfig()
    exitKeys, keyExitKeysStr, keyUp, keyDown, keyFastUp, keyFastDown, keyEnter = cfg.GetKeys()
    keyMap = map[string]string{
        keyUp:            "↑",
        keyDown:          "↓",
        keyFastDown:      "pgDown",
        keyFastUp:        "pgUp",
        keyEnter:         "⏎",
    }
)