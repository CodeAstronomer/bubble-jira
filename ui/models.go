package ui

import (
	"bubble-jira/config"
	"bubble-jira/jira"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	configList       configListEditor
	configInput      configInputEditor
	fetching         fetchingModel
	commentsViewport viewport.Model
	statusBar        string
	editingFieldIdx  int
	configValidError string
	licenceContent   string
	licenceLoading   bool
	licenceOffset    int
	comments         []jira.Comment
	commentsLoading  bool
	screenWidth      int
	screenHeight     int
	editingMode      string
	hoverTimer       *time.Timer
    hoveredTaskKey   string
    cachedComments   map[string][]jira.Comment
    fetchingComments bool
    statusMessage    string
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
    fetchCommentsAfterProgress bool
}

// fetchingModel represents the fetching state UI
type fetchingModel struct {
	spinner      spinner.Model
	progress     progress.Model
	stages       []string
	currentStage int
	fromStatus   bool
	status       string
	error        string
	done         bool
	width        int
}

// configListEditor handles config list editing
type configListEditor struct {
	fields []configField
	list   list.Model
}

// configField represents a configuration field
type configField struct {
	key   string
	value string
}

func (cf configField) Title() string       { return cf.key }
func (cf configField) Description() string { return cf.value }
func (cf configField) FilterValue() string { return cf.key }

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

// contextMenuItem represents a context menu item
type contextMenuItem struct {
	title string
}

func (cm contextMenuItem) Title() string       { return cm.title }
func (cm contextMenuItem) Description() string { return "" }
func (cm contextMenuItem) FilterValue() string { return cm.title }

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

	vp := viewport.New(200, 50)
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
	switch msg.(type) {
	case startTasksMsg:
		m.state = "fetching"
		m.fetching = newFetchingModel()
		m.configValidError = ""
		return m, tea.Batch(
			fetchJiraTasksCmd(m.jc),
			tickFetchCmd(),
			m.fetching.spinner.Tick,
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
	default:
		return m, nil
	}
}

// View renders the current view based on state
func (m model) View() string {
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
		return menuStyle.Render(m.licenceView())
	case "tasks":
		return tasksStyle.Render(m.tasksTableView())
	case "task-context":
		return tasksStyle.Render(m.taskContextView())
	case "comments-fetching":
		return fetchingStyle.Render(m.fetchingView())
	case "comments-view":
		return m.commentsView()
	case "config":
		return configStyle.Render(m.configListView())
	case "config-edit":
		return configStyle.Render(m.configInputView())
	case "fetching":
		return fetchingStyle.Render(m.fetchingView())
    case "task-status":
        return fetchingStyle.Render(m.taskStatusView())
    case "task-settings-list":
        return menuStyle.Render(m.taskSettings.View())
    case "add-comment":
        return fetchingStyle.Render(m.addCommentView())
	default:
		return "Unknown state"
	}
}