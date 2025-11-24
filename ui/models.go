package ui

import (
	"bubble-jira/config"
	"bubble-jira/jira-code"
	"bubble-jira/ui/views"
	"bubble-jira/ui/types"
	"time"

	"github.com/charmbracelet/bubbles/list"


	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type startTasksMsg struct{}
type quitAfterDelayMsg struct{}

// model represents the main application model
type model struct {
	cfg              *config.Config
	jc               *jira.Client
	state            string
	menu             list.Model
	settings         list.Model
	tasksTable       table.Model
	taskContextMenu  list.Model
	taskSettings     list.Model
	selectedIssue    *jira.Issue
	configList       types.ConfigListEditor
	configInput      configInputEditor
	fetching         types.FetchingModel
	commentsViewport viewport.Model
	statusBar        string
	editingFieldIdx  int
	configValidError string
	licenceContent   string
	licenceLoading   bool
	licenceOffset    int
	comments         []jira.Comment
	issueDescription string
	commentsLoading  bool
	screenWidth      int
	screenHeight     int
	editingMode      string
	hoverTimer       *time.Timer
    hoveredTaskKey   string
    cachedComments   map[string][]jira.Comment
    cachedDescriptions map[string]string
    fetchingComments bool
    statusMessage    string
    isFromAddComments bool
    commitInput struct {
    	input     textinput.Model
        focusSave bool
        key       string
        title     string
    }
    jiraStatusInput struct {
    	key      string
        input    textinput.Model
        focusSave bool
        cursor   int
        choice   string
    }
    addCommentInput struct {
        input     textinput.Model
        focusSend bool
        issueKey  string
    }
    commitGitMessage struct {
        input     textinput.Model
        focusSend bool
        issueKey  string
    }
    gitIssueLoc struct {
        input    textinput.Model
        focusSave bool
        cursor   int
        choice   string
    }
    gitIssueStyle struct {
        input    textinput.Model
        focusSave bool
        cursor   int
        choice   string
    }
    ChooseLanguage struct {
        input    textinput.Model
        focusSave bool
        cursor   int
        choice   string
    }
    taskSearchInput  textinput.Model
    originalTaskRows []table.Row
}





// configInputEditor handles individual config field editing
type configInputEditor struct {
	input     textinput.Model
	key       string
	focusSave bool
}

// menuItem represents a menu item
type menuItem struct {
	title   string
	enabled bool
}

func (m menuItem) Title() string       { return m.title }
func (m menuItem) Description() string { return "" }
func (m menuItem) FilterValue() string { return m.title }



// newModel creates the initial model state
func newModel(cfg *config.Config, jc *jira.Client) model {
	menuItems := []list.Item{
    	menuItem{title: Strings["MenuViewTasksTitle"], enabled: cfg.IsValid()},
    	menuItem{title: Strings["MenuSettingsTitle"], enabled: true},
    	menuItem{title: Strings["MenuQuitTitle"], enabled: true},
    }
	menu := list.New(menuItems, list.NewDefaultDelegate(), terminalWidth, terminalHeight)
	menu.Title = Strings["menuTitle"]
	menu.SetShowHelp(true)
	menu.SetShowPagination(false)


    footer := lipgloss.NewStyle().
        Faint(true).
        Render(keyMap[keyUp] + "/" + keyMap[keyDown] + ": scroll • " + keyNewComment + ": Add Comment • " + keyExitKeysStr + ": back")
    footerHeight := lipgloss.Height(footer)
	vp := viewport.New(terminalWidth, terminalHeight-footerHeight)
	vp.Style = commentsStyle

	return model{
		cfg:              cfg,
		jc:               jc,
		state:            "menu",
		menu:             menu,
		fetching:         newFetchingModel(),
		commentsViewport: vp,
		screenWidth:      terminalWidth,
		screenHeight:     terminalHeight,
		cachedComments:   make(map[string][]jira.Comment),
		cachedDescriptions: make(map[string]string),
        fetchingComments: false,
        hoverTimer:       nil,
        hoveredTaskKey:   "",
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles all user input and messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case keyForceQuit:
            return m, tea.Quit
        }
    }

	switch msg.(type) {
	case startTasksMsg:
		m.state = "fetching"
		m.fetching = newFetchingModel()
		m.configValidError = ""
		return m, tea.Batch(
			fetchJiraTasksCmd(m.jc),
			tickFetchCmd(),
			m.fetching.Spinner.Tick,
		)
	}

    // fallback to normal update logic
	switch m.state {
	case "menu":
		return m.updateMenu(msg)
	case "settings":
		return m.updateSettings(msg)
	case "licence":
		return m.updateLicence(msg)
	case "fetching":
		return m.updateFetching(msg)
	case "tasks":
		return m.updateTasks(msg)
	case "task-context":
		return m.updateTaskContext(msg)
	case "comments-fetching":
		return m.updateCommentsFetching(msg)
	case "comments-view":
		return m.updateCommentsView(msg)
	case "config":
		return m.updateConfigList(msg)
	case "config-edit":
		return m.updateConfigEdit(msg)
	case "task-status":
        return m.updateTaskStatus(msg)
	case "task-settings-list":
        return m.updateTaskSettings(msg)
    case "add-comment":
        return m.updateAddCommentInput(msg)
    case "enter-commit-message":
        return m.updateEnterCommitMessage(msg)
    case "task-search":
        return m.updateTaskSearch(msg)
    case "issue-git-location":
        return m.updateIssueGitLocation(msg)
    case "issue-git-style":
        return m.updateIssueGitStyle(msg)
    case "select-Language":
        return m.updateChooseLanguage(msg)
	default:
		return m, nil
	}
}

// View renders the current view based on state
func (m model) View() string {
    keyData := types.KeyData{
        Up:       keyUp,
        Down:     keyDown,
        FastUp:   keyFastUp,
        FastDown: keyFastDown,
        Exit:     keyExitKeysStr,
		Enter:    keyEnter,
		Search:   keySearch,
		Comment:  keyNewComment,
    }

	switch m.state {
	case "menu":
		view := m.menu.View()
		if m.configValidError != "" {
			view = view + "\n\n" + errorStyle.Render(m.configValidError)
		}
		return menuStyle.Render(view)
	case "settings":
		return menuStyle.Render(m.settings.View())
	case "licence":
		viewData := types.LicenseViewData{
			Content:      m.licenceContent,
			IsLoading:    m.licenceLoading,
			Offset:       m.licenceOffset,
			ScreenHeight: m.screenHeight,
			ScreenWidth:  m.screenWidth,
			Keys:         keyData,
			Strings:      Strings,
		}
		return menuStyle.Render(views.LicenseView(viewData))
	case "tasks":
		viewData := types.TasksListData{
			TasksTable:     m.tasksTable,
			TableBaseStyle: tableBaseStyle,
			Keys:           keyData,
			Strings:        Strings,
		}
		return tasksStyle.Render(views.TasksTableView(viewData))
	case "task-context":
		viewData := types.TaskContextViewData{
			Menu: m.taskContextMenu,
		}
		return tasksStyle.Render(views.TaskContextView(viewData))
	case "comments-fetching":
		return fetchingStyle.Render(views.FetchingView(m.fetching))
	case "comments-view":
        viewData := types.CommentViewData{
            CommentsViewPort:        m.commentsViewport.View(),
			IsLoading:    m.commentsLoading,
			ScreenHeight: m.screenHeight,
			ScreenWidth:  m.screenWidth,
			Keys:         keyData,
			Strings:      Strings,
		}
		return menuStyle.Render(views.CommentsView(viewData))
	case "config":
		return configStyle.Render(views.ConfigListView(m.configList))
	case "config-edit":
		return configStyle.Render(m.configInputView())
	case "fetching":
		return fetchingStyle.Render(views.FetchingView(m.fetching))
    case "task-status":
        viewData := types.TaskStatusViewData{
            ScreenHeight:    m.screenHeight,
            ScreenWidth:     m.screenWidth,
            Keys:            keyData,
            Strings:         Strings,
            Statuses:        statuses,
			JiraStatusInput: m.jiraStatusInput.cursor,
        }
        return menuStyle.Render(views.TaskStatusView(viewData))
    case "task-settings-list":
        return menuStyle.Render(m.taskSettings.View())
    case "add-comment":
        viewData := types.AddCommentViewData{
            ScreenHeight:        m.screenHeight,
            ScreenWidth:         m.screenWidth,
            Keys:                keyData,
            Strings:             Strings,
            AddCommentInput:     m.addCommentInput.focusSend,
            AddCommentInputView: m.addCommentInput.input.View(),
            FocusedButton:       focusedButton,
            BlurredButton:       blurredButton,
        }
        return menuStyle.Render(views.AddCommentView(viewData))
    case "enter-commit-message":
        viewData := types.EnterCommitMessageData{
            ScreenHeight:           m.screenHeight,
            ScreenWidth:            m.screenWidth,
            Keys:                   keyData,
            Strings:                Strings,
            CommitGitMessage:       m.commitGitMessage.focusSend,
            CommitGitMessageView:   m.commitGitMessage.input.View(),
            FocusedButtonGit:       focusedButtonGit,
            BlurredButtonGit:       blurredButtonGit,
        }
        return menuStyle.Render(views.EnterCommitMessage(viewData))
    case "task-search":
		viewData := types.TaskSearchViewData{
			TaskSearchInput: m.taskSearchInput,
			TasksTable:      m.tasksTable,
			TableBaseStyle:  tableBaseStyle,
			Strings:         Strings,
			KeyMap:          keyMap,
			KeyEnter:        keyEnter,
		}
		return views.TaskSearchView(viewData)
    case "issue-git-location":
        return m.issueGitLocation()
    case "issue-git-style":
        return m.issueGitStyle()
    case "select-Language":
        return m.selectLanguage()
	default:
		return "Unknown state"
	}
}