package app

import "github.com/charmbracelet/lipgloss"

var HeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7"))

var IdeaInputCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

var IdeaInputCursorLineStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230"))

var IdeaInputBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62"))

var WrittenIdeasViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("238")).
	PaddingRight(2)

var RankingIdeaViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	PaddingRight(2)

var GoodIdeaViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#FF06B7")).
	PaddingRight(2)

var BadIdeaViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("238")).
	PaddingRight(2)

var HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
