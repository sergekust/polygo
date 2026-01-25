package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7"))

var ideaInputCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

var ideaInputCursorLineStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230"))

var ideaInputBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62"))

var writtenIdeasViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("238")).
	PaddingRight(2)

var rankingIdeaViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	PaddingRight(2)

var goodIdeaViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#FF06B7")).
	PaddingRight(2)

var badIdeaViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("238")).
	PaddingRight(2)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type model struct {
	timer timer.Model

	// Timer Settings
	minutesInput textinput.Model
	secondsInput textinput.Model
	focused      string

	// Ideas
	ideas        []string
	ideaInput    textarea.Model
	writtenIdeas viewport.Model

	// Ranking
	ideaRank            map[int][]int
	rankingIdea         int
	rankingIdeaViewport viewport.Model
	goodIdeasViewport   viewport.Model
	badIdeasViewport    viewport.Model
}

func initialModel() model {
	minutesInput := textinput.New()
	minutesInput.Placeholder = "00"
	minutesInput.CharLimit = 2
	minutesInput.Width = 2
	minutesInput.Prompt = ""
	minutesInput.Focus()

	secondsInput := textinput.New()
	secondsInput.Placeholder = "00"
	secondsInput.CharLimit = 2
	secondsInput.Width = 2
	secondsInput.Prompt = ""
	secondsInput.Blur()

	ideaInput := textarea.New()
	ideaInput.Placeholder = ""
	ideaInput.SetHeight(15)
	ideaInput.SetWidth(75)
	ideaInput.Blur()
	ideaInput.FocusedStyle.CursorLine = ideaInputCursorLineStyle
	ideaInput.FocusedStyle.Base = ideaInputBorderStyle
	ideaInput.Cursor.Style = ideaInputCursorStyle

	writtenIdeaViewport := viewport.New(75, 17)
	writtenIdeaViewport.Style = writtenIdeasViewportStyle

	rankingIdeaViewport := viewport.New(50, 17)
	rankingIdeaViewport.Style = rankingIdeaViewportStyle

	goodIdeasViewport := viewport.New(50, 17)
	goodIdeasViewport.Style = goodIdeaViewportStyle

	badIdeasViewport := viewport.New(50, 17)
	badIdeasViewport.Style = badIdeaViewportStyle

	return model{
		focused:             "minutes", // minutes, seconds, idea, ranking, store
		minutesInput:        minutesInput,
		secondsInput:        secondsInput,
		ideaInput:           ideaInput,
		writtenIdeas:        writtenIdeaViewport,
		ideaRank:            map[int][]int{1: make([]int, 0), 2: make([]int, 0)},
		goodIdeasViewport:   goodIdeasViewport,
		badIdeasViewport:    badIdeasViewport,
		rankingIdeaViewport: rankingIdeaViewport,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.focused = "ranking"

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit

		case "left":
			if m.focused == "ranking" {
				m.ideaRank[1] = append(m.ideaRank[1], m.rankingIdea)
				m.rankingIdea++
				if m.rankingIdea >= len(m.ideas) {
					m.focused = "store"
					m.StoreIndeasIntoFile()
				}
			}

		case "right":
			if m.focused == "ranking" {
				m.ideaRank[2] = append(m.ideaRank[2], m.rankingIdea)
				m.rankingIdea++
				if m.rankingIdea >= len(m.ideas) {
					m.focused = "store"
					m.StoreIndeasIntoFile()
				}
			}

		case "enter":
			if m.focused == "minutes" {
				m.focused = "seconds"
				m.minutesInput.Blur()
				m.secondsInput.Focus()
				return m, nil
			}

			if m.focused == "seconds" {
				// Stop listening timer settings input
				m.minutesInput.Blur()
				m.secondsInput.Blur()

				// Set timer timeout
				minutesValue, err := strconv.Atoi(m.minutesInput.Value())
				if err != nil {
					minutesValue = 0
				}
				secondsValue, err := strconv.Atoi(m.secondsInput.Value())
				if err != nil {
					secondsValue = 0
				}
				timeout := time.Second * time.Duration(secondsValue+(minutesValue*60))

				// Set timer
				m.timer = timer.NewWithInterval(timeout, time.Second)
				cmd := m.timer.Init()

				// Start listening idea input
				m.focused = "idea"
				m.ideaInput.Focus()
				return m, cmd
			}

		case "tab":
			if m.focused == "minutes" {
				m.focused = "seconds"
				m.minutesInput.Blur()
				m.secondsInput.Focus()
				return m, nil
			}
			if m.focused == "seconds" {
				m.focused = "minutes"
				m.secondsInput.Blur()
				m.minutesInput.Focus()
				return m, nil
			}
			if m.focused == "idea" {
				clearedInput := strings.TrimSpace(m.ideaInput.Value())
				clearedInput = strings.TrimSuffix(clearedInput, "\n")

				if clearedInput == "" {
					return m, nil
				}
				m.ideas = append(m.ideas, fmt.Sprintf("\n%s\n\n", clearedInput))
				m.ideaInput.Reset()
				return m, nil
			}
		}
	}

	if m.ideaInput.Focused() {
		cmds := make([]tea.Cmd, 2)
		m.ideaInput, cmds[0] = m.ideaInput.Update(msg)
		m.writtenIdeas, cmds[1] = m.writtenIdeas.Update(msg)
		return m, tea.Batch(cmds...)
	}

	if m.minutesInput.Focused() || m.secondsInput.Focused() {
		cmds := make([]tea.Cmd, 2)
		m.minutesInput, cmds[0] = m.minutesInput.Update(msg)
		m.secondsInput, cmds[1] = m.secondsInput.Update(msg)
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m *model) UpdateWrittenIdeasViewport() {
	str, _ := glamour.Render(strings.Join(m.ideas, "---"), "dark")
	m.writtenIdeas.SetContent(str)
	m.writtenIdeas.GotoBottom()
}

func (m model) StoreIndeasIntoFile() {
	if len(m.ideaRank[1]) == 0 && len(m.ideaRank[2]) == 0 {
		return
	}

	var sb strings.Builder

	if len(m.ideaRank[1]) > 0 {
		sb.WriteString("# Fav ideas")
		for _, v := range m.ideaRank[1] {
			sb.WriteString(fmt.Sprintf("%s---", m.ideas[v]))
		}
	}

	if len(m.ideaRank[2]) > 0 {
		sb.WriteString("\n\n# Ideas to be polished")
		for _, v := range m.ideaRank[2] {
			sb.WriteString(fmt.Sprintf("%s---", m.ideas[v]))
		}
	}

	filename := "IDEAS.md"
	permissions := os.FileMode(0644)
	os.WriteFile(filename, []byte(sb.String()), permissions)
	sb.Reset()
}

func (m *model) UpdateRanking() {
	// Update good ideas
	goodIdeas := make([]string, 0, 2)
	for _, v := range m.ideaRank[1] {
		goodIdeas = append(goodIdeas, m.ideas[v])
	}

	goodIdeasContent, _ := glamour.Render(strings.Join(goodIdeas, "---"), "dark")
	m.goodIdeasViewport.SetContent(goodIdeasContent)
	m.goodIdeasViewport.GotoBottom()

	// Update current
	rankingIdeaContent, _ := glamour.Render(m.ideas[m.rankingIdea], "dark")
	m.rankingIdeaViewport.SetContent(rankingIdeaContent)
	m.rankingIdeaViewport.GotoTop()

	// Update bad ideas
	badIdeas := make([]string, 0, 2)
	for _, v := range m.ideaRank[2] {
		badIdeas = append(badIdeas, m.ideas[v])
	}
	badIdeasContent, _ := glamour.Render(strings.Join(badIdeas, "---"), "dark")
	m.badIdeasViewport.SetContent(badIdeasContent)
	m.badIdeasViewport.GotoBottom()
}

func (m model) View() string {
	// Header
	s := headerStyle.Render("//// Polygo /////////////////////////////")
	s += "\n\n"

	if m.timer.Running() {
		s += fmt.Sprintf("Time left: %s\n\n", m.timer.View())

		nextIdeaView := m.ideaInput.View()

		m.UpdateWrittenIdeasViewport()
		storedIdeasView := m.writtenIdeas.View()

		s += lipgloss.JoinHorizontal(lipgloss.Top, nextIdeaView, storedIdeasView)

		s += helpStyle("\n\n⇄ TAB - store an idea")
	}

	if m.focused == "minutes" || m.focused == "seconds" {
		s += fmt.Sprintf(
			"Set timer:\n%s: %s",
			m.minutesInput.View(),
			m.secondsInput.View(),
		)

		s += helpStyle("\n\n⇄ TAB - jump between minutes and seconds\n⏎ ENTER - start a session\nCtrl+C - exit")
	}

	if m.focused == "ranking" {
		s += "Whoa, brain freeze!\nTimer's buzzed—time to sort ideas. Swipe left for the fav ones, right for the nahs.\nLet's make magic! ✨\n\n"

		m.UpdateRanking()
		goodIdeasView := m.goodIdeasViewport.View()
		currentIdeaView := m.rankingIdeaViewport.View()
		badIdeasView := m.badIdeasViewport.View()

		s += lipgloss.JoinHorizontal(lipgloss.Top, goodIdeasView, currentIdeaView, badIdeasView)
		s += helpStyle("\n\n← LEFT - like\n→ RIGHT - need polishing\nCtrl+C - exit")
	}

	if m.focused == "store" {
		confirmation, _ := glamour.Render("# DONE\n\nFile `IDEAS.md` is saved!", "dark")
		s += confirmation
		s += helpStyle("\n\nCtrl+C - exit")
	}

	return s
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}
