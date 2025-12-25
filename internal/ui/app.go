package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kenan/xjson/internal/api"
	"github.com/kenan/xjson/internal/transform"
)

// View modes
type viewMode int

const (
	viewTimeline viewMode = iota
	viewProfile
	viewSearch
)

// App is the main application model
type App struct {
	client   *api.Client
	keys     KeyMap
	help     help.Model
	viewport viewport.Model
	input    textinput.Model

	// State
	mode          viewMode
	ready         bool
	searching     bool
	loading       bool
	err           error
	width         int
	height        int

	// Data
	timeline      *transform.DisguisedResponse
	profile       *transform.DisguisedPayload
	searchResults *transform.DisguisedResponse
	currentIndex  int
	nextToken     string

	// Display
	jsonContent   string
	statusLine    string
}

// NewApp creates a new application instance
func NewApp(client *api.Client) *App {
	ti := textinput.New()
	ti.Placeholder = "search query..."
	ti.CharLimit = 256

	return &App{
		client:     client,
		keys:       DefaultKeyMap(),
		help:       help.New(),
		input:      ti,
		statusLine: "Initializing...",
	}
}

// Message types
type (
	timelineMsg    *transform.DisguisedResponse
	profileMsg     *transform.DisguisedPayload
	searchMsg      *transform.DisguisedResponse
	errMsg         error
)

// Init initializes the app
func (a *App) Init() tea.Cmd {
	return a.fetchTimeline()
}

// fetchTimeline fetches the home timeline
func (a *App) fetchTimeline() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := a.client.GetHomeTimeline(ctx, 20, a.nextToken)
		if err != nil {
			return errMsg(err)
		}

		disguised := transform.TransformTimeline(resp, "/2/timeline/home")
		if resp.Meta != nil {
			a.nextToken = resp.Meta.NextToken
		}

		return timelineMsg(disguised)
	}
}

// fetchProfile fetches a user profile
func (a *App) fetchProfile(username string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		user, err := a.client.GetUser(ctx, username)
		if err != nil {
			return errMsg(err)
		}

		disguised := transform.TransformUser(user)
		return profileMsg(&disguised)
	}
}

// searchTweets searches for tweets
func (a *App) searchTweets(query string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := a.client.SearchTweets(ctx, query, 20, "")
		if err != nil {
			return errMsg(err)
		}

		disguised := transform.TransformSearch(resp, query)
		return searchMsg(disguised)
	}
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		headerHeight := 3  // Title + request line
		footerHeight := 2  // Help line

		if !a.ready {
			a.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			a.viewport.YPosition = headerHeight
			// Configure viewport keys
			a.viewport.KeyMap = viewport.KeyMap{
				Up:       key.NewBinding(key.WithKeys("up", "k")),
				Down:     key.NewBinding(key.WithKeys("down", "j")),
				PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+u")),
				PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+d")),
				HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u")),
				HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d")),
			}
			a.ready = true
		} else {
			a.viewport.Width = msg.Width
			a.viewport.Height = msg.Height - headerHeight - footerHeight
		}

		a.updateContent()

	case tea.KeyMsg:
		if a.searching {
			switch {
			case msg.String() == "enter":
				a.searching = false
				query := a.input.Value()
				if query != "" {
					a.loading = true
					a.statusLine = fmt.Sprintf("GET /2/tweets/search/recent?q=%s...", query)
					return a, a.searchTweets(query)
				}
			case msg.String() == "esc":
				a.searching = false
				a.input.Reset()
			default:
				var cmd tea.Cmd
				a.input, cmd = a.input.Update(msg)
				return a, cmd
			}
			return a, nil
		}

		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit

		case key.Matches(msg, a.keys.Next):
			a.nextItem()
			a.updateContent()
			return a, nil

		case key.Matches(msg, a.keys.Prev):
			a.prevItem()
			a.updateContent()
			return a, nil

		case key.Matches(msg, a.keys.Refresh):
			a.loading = true
			a.currentIndex = 0
			a.nextToken = ""
			a.statusLine = "GET /2/timeline/home..."
			return a, a.fetchTimeline()

		case key.Matches(msg, a.keys.Search):
			a.searching = true
			a.input.Focus()
			return a, textinput.Blink

		case key.Matches(msg, a.keys.Timeline):
			a.mode = viewTimeline
			a.loading = true
			a.currentIndex = 0
			a.statusLine = "GET /2/timeline/home..."
			return a, a.fetchTimeline()

		case key.Matches(msg, a.keys.Help):
			a.help.ShowAll = !a.help.ShowAll
			return a, nil
		}

		// Pass other keys to viewport for scrolling
		var cmd tea.Cmd
		a.viewport, cmd = a.viewport.Update(msg)
		return a, cmd

	case timelineMsg:
		a.loading = false
		a.mode = viewTimeline
		a.timeline = msg
		a.statusLine = fmt.Sprintf("GET %s - 200 OK (%sms)", msg.Endpoint, msg.Latency)
		a.updateContent()

	case profileMsg:
		a.loading = false
		a.mode = viewProfile
		a.profile = msg
		a.statusLine = fmt.Sprintf("GET %s - 200 OK", msg.Endpoint)
		a.updateContent()

	case searchMsg:
		a.loading = false
		a.mode = viewSearch
		a.searchResults = msg
		a.statusLine = fmt.Sprintf("GET %s - 200 OK (%sms)", msg.Endpoint, msg.Latency)
		a.updateContent()

	case errMsg:
		a.loading = false
		a.err = msg
		a.statusLine = fmt.Sprintf("Error: %v", msg)
	}

	return a, tea.Batch(cmds...)
}

// nextItem moves to the next item
func (a *App) nextItem() {
	var maxIndex int
	switch a.mode {
	case viewTimeline:
		if a.timeline != nil {
			maxIndex = len(a.timeline.Data) - 1
		}
	case viewSearch:
		if a.searchResults != nil {
			maxIndex = len(a.searchResults.Data) - 1
		}
	}

	if a.currentIndex < maxIndex {
		a.currentIndex++
	}
}

// prevItem moves to the previous item
func (a *App) prevItem() {
	if a.currentIndex > 0 {
		a.currentIndex--
	}
}

// updateContent updates the viewport content
func (a *App) updateContent() {
	var content string
	var err error

	switch a.mode {
	case viewTimeline:
		if a.timeline != nil && len(a.timeline.Data) > 0 {
			content, err = transform.ToJSON(a.timeline.Data[a.currentIndex])
		}
	case viewProfile:
		if a.profile != nil {
			content, err = transform.ToJSON(a.profile)
		}
	case viewSearch:
		if a.searchResults != nil && len(a.searchResults.Data) > 0 {
			content, err = transform.ToJSON(a.searchResults.Data[a.currentIndex])
		}
	}

	if err != nil {
		content = fmt.Sprintf("Error rendering JSON: %v", err)
	}

	a.jsonContent = highlightJSON(content)
	a.viewport.SetContent(a.jsonContent)
}

// highlightJSON applies syntax highlighting to JSON
func highlightJSON(s string) string {
	var result strings.Builder
	inString := false
	afterColon := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		switch {
		case c == '"':
			if i > 0 && s[i-1] == '\\' {
				result.WriteByte(c)
				continue
			}

			if inString {
				result.WriteByte(c)
				result.WriteString("\033[0m") // Reset
				inString = false
				afterColon = false
			} else {
				inString = true
				// Check if this is a key (followed eventually by :)
				isKey := false
				for j := i + 1; j < len(s); j++ {
					if s[j] == '"' {
						// Look for colon after the string
						for k := j + 1; k < len(s); k++ {
							if s[k] == ':' {
								isKey = true
								break
							} else if s[k] != ' ' && s[k] != '\n' && s[k] != '\t' {
								break
							}
						}
						break
					}
				}

				if isKey {
					result.WriteString("\033[38;5;203m") // Bright coral/red for keys
				} else {
					result.WriteString("\033[38;5;114m") // Soft green for strings
				}
				result.WriteByte(c)
			}

		case c == ':' && !inString:
			result.WriteString("\033[38;5;245m") // Gray colon
			result.WriteByte(c)
			result.WriteString("\033[0m")
			afterColon = true

		case (c >= '0' && c <= '9') || c == '-' || c == '.':
			if !inString && afterColon {
				result.WriteString("\033[38;5;215m") // Orange for numbers
				result.WriteByte(c)
				// Continue reading number
				for i+1 < len(s) && ((s[i+1] >= '0' && s[i+1] <= '9') || s[i+1] == '.' || s[i+1] == 'e' || s[i+1] == 'E' || s[i+1] == '+' || s[i+1] == '-') {
					i++
					result.WriteByte(s[i])
				}
				result.WriteString("\033[0m")
				afterColon = false
			} else {
				result.WriteByte(c)
			}

		case c == 't' && !inString && i+3 < len(s) && s[i:i+4] == "true":
			result.WriteString("\033[38;5;79m") // Teal for true
			result.WriteString("true")
			result.WriteString("\033[0m")
			i += 3
			afterColon = false

		case c == 'f' && !inString && i+4 < len(s) && s[i:i+5] == "false":
			result.WriteString("\033[38;5;204m") // Pink for false
			result.WriteString("false")
			result.WriteString("\033[0m")
			i += 4
			afterColon = false

		case c == 'n' && !inString && i+3 < len(s) && s[i:i+4] == "null":
			result.WriteString("\033[38;5;139m") // Purple for null
			result.WriteString("null")
			result.WriteString("\033[0m")
			i += 3
			afterColon = false

		case c == '{' || c == '}':
			result.WriteString("\033[38;5;222m\033[1m") // Bold gold for braces
			result.WriteByte(c)
			result.WriteString("\033[0m")
			if c == '{' {
				afterColon = false
			}

		case c == '[' || c == ']':
			result.WriteString("\033[38;5;147m\033[1m") // Bold lavender for brackets
			result.WriteByte(c)
			result.WriteString("\033[0m")
			if c == '[' {
				afterColon = false
			}

		case c == ',' && !inString:
			result.WriteByte(c)
			afterColon = false

		default:
			result.WriteByte(c)
		}
	}

	return result.String()
}

// View renders the app
func (a *App) View() string {
	if !a.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Title bar
	title := TitleStyle.Width(a.width).Render("API Response Inspector v1.0.0")
	b.WriteString(title)
	b.WriteString("\n")

	// Request/status line
	var statusStyle lipgloss.Style
	if a.err != nil {
		statusStyle = ErrorRequestStyle
	} else {
		statusStyle = RequestStyle
	}

	statusText := a.statusLine
	if a.loading {
		statusText += " [Loading...]"
	}

	// Add item counter for list views
	if a.mode == viewTimeline && a.timeline != nil && len(a.timeline.Data) > 0 {
		statusText = fmt.Sprintf("%s  [%d/%d]", statusText, a.currentIndex+1, len(a.timeline.Data))
	} else if a.mode == viewSearch && a.searchResults != nil && len(a.searchResults.Data) > 0 {
		statusText = fmt.Sprintf("%s  [%d/%d]", statusText, a.currentIndex+1, len(a.searchResults.Data))
	}

	status := statusStyle.Width(a.width).Render(statusText)
	b.WriteString(status)
	b.WriteString("\n")

	// Search input (if active)
	if a.searching {
		searchLine := SearchStyle.Render("Search: ") + a.input.View()
		b.WriteString(searchLine)
		b.WriteString("\n")
	}

	// Main content
	b.WriteString(a.viewport.View())
	b.WriteString("\n")

	// Help bar
	helpView := a.help.View(a.keys)
	b.WriteString(HelpStyle.Width(a.width).Render(helpView))

	return b.String()
}
