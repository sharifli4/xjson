package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors - vibrant developer-tool colors
	primaryColor   = lipgloss.Color("#7AA2F7") // Soft blue
	secondaryColor = lipgloss.Color("#9ECE6A") // Bright green
	errorColor     = lipgloss.Color("#F7768E") // Bright pink/red
	warningColor   = lipgloss.Color("#FF9E64") // Orange
	mutedColor     = lipgloss.Color("#565F89") // Muted purple
	bgColor        = lipgloss.Color("#1A1B26") // Dark bg
	fgColor        = lipgloss.Color("#C0CAF5") // Light purple-white

	// JSON syntax colors
	jsonKeyColor    = lipgloss.Color("#E06C75")
	jsonStringColor = lipgloss.Color("#98C379")
	jsonNumberColor = lipgloss.Color("#D19A66")
	jsonBoolColor   = lipgloss.Color("#56B6C2")
	jsonNullColor   = lipgloss.Color("#C678DD")

	// Title bar style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282C34")).
			Background(primaryColor).
			Padding(0, 1)

	// Status bar style
	StatusStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(lipgloss.Color("#21252B")).
			Padding(0, 1)

	// Request info style (GET /endpoint - 200 OK)
	RequestStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Background(lipgloss.Color("#21252B")).
			Padding(0, 1)

	// Error request style
	ErrorRequestStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Background(lipgloss.Color("#21252B")).
				Padding(0, 1)

	// Main content area
	ContentStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Padding(1, 2)

	// Help bar at bottom
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1)

	// Selected item highlight
	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3E4451")).
			Foreground(fgColor)

	// Search input
	SearchStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Padding(0, 1)

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor)
)

// JSON syntax highlighting helpers
func StyleJSONKey(s string) string {
	return lipgloss.NewStyle().Foreground(jsonKeyColor).Render(s)
}

func StyleJSONString(s string) string {
	return lipgloss.NewStyle().Foreground(jsonStringColor).Render(s)
}

func StyleJSONNumber(s string) string {
	return lipgloss.NewStyle().Foreground(jsonNumberColor).Render(s)
}

func StyleJSONBool(s string) string {
	return lipgloss.NewStyle().Foreground(jsonBoolColor).Render(s)
}

func StyleJSONNull(s string) string {
	return lipgloss.NewStyle().Foreground(jsonNullColor).Render(s)
}

func StyleJSONBracket(s string) string {
	return lipgloss.NewStyle().Foreground(fgColor).Render(s)
}

func StyleMethod(method string) string {
	color := secondaryColor
	switch method {
	case "GET":
		color = secondaryColor
	case "POST":
		color = warningColor
	case "DELETE":
		color = errorColor
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(method)
}

func StyleStatusCode(code int) string {
	color := secondaryColor
	if code >= 400 {
		color = errorColor
	} else if code >= 300 {
		color = warningColor
	}
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%d", code))
}

