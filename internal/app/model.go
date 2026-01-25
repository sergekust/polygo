package app

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

type Model struct {
	// Timer Settings
	startedAt    time.Time
	timer        timer.Model
	minutesInput textinput.Model
	secondsInput textinput.Model
	focused      string

	// Ideas
	ideasStorage IdeaStrorage
	ideaInput    textarea.Model
	writtenIdeas viewport.Model

	// Ranking
	rankingIdeaViewport viewport.Model
	goodIdeasViewport   viewport.Model
	badIdeasViewport    viewport.Model

	// File with result
	resultFilename string
}

func NewModel() Model {
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
	ideaInput.FocusedStyle.CursorLine = IdeaInputCursorLineStyle
	ideaInput.FocusedStyle.Base = IdeaInputBorderStyle
	ideaInput.Cursor.Style = IdeaInputCursorStyle

	writtenIdeaViewport := viewport.New(75, 17)
	writtenIdeaViewport.Style = WrittenIdeasViewportStyle

	rankingIdeaViewport := viewport.New(50, 17)
	rankingIdeaViewport.Style = RankingIdeaViewportStyle

	goodIdeasViewport := viewport.New(50, 17)
	goodIdeasViewport.Style = GoodIdeaViewportStyle

	badIdeasViewport := viewport.New(50, 17)
	badIdeasViewport.Style = BadIdeaViewportStyle

	return Model{
		focused:             "minutes", // minutes, seconds, idea, ranking, store
		minutesInput:        minutesInput,
		secondsInput:        secondsInput,
		ideaInput:           ideaInput,
		writtenIdeas:        writtenIdeaViewport,
		goodIdeasViewport:   goodIdeasViewport,
		badIdeasViewport:    badIdeasViewport,
		rankingIdeaViewport: rankingIdeaViewport,
		resultFilename:      "IDEAS.md",
		ideasStorage:        NewIdeaStorage(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.ideasStorage.RankCurrentIdea(true)
				if m.ideasStorage.AreAllIdeasRanked() {
					m.focused = "store"
					m.storeIdeasIntoFile()
				}
			}

		case "right":
			if m.focused == "ranking" {
				m.ideasStorage.RankCurrentIdea(false)
				if m.ideasStorage.AreAllIdeasRanked() {
					m.focused = "store"
					m.storeIdeasIntoFile()
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
				m.startedAt = time.Now()
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
				m.ideasStorage.Add(fmt.Sprintf("\n%s\n\n", clearedInput))
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

func (m *Model) updateWrittenIdeasViewport() {
	str, _ := glamour.Render(strings.Join(m.ideasStorage.ideas, "---"), "dark")
	m.writtenIdeas.SetContent(str)
	m.writtenIdeas.GotoBottom()
}

func (m Model) storeIdeasIntoFile() {
	if len(m.ideasStorage.ideas) == 0 {
		return
	}

	var sb strings.Builder

	// Write session info
	sb.WriteString(
		fmt.Sprintf("Session started: `%s`", m.startedAt.Format("2006-01-02 15:04:05")),
	)
	sb.WriteString(
		fmt.Sprintf("\nSession closed: `%s`", time.Now().Format("2006-01-02 15:04:05")),
	)

	// Write ideas
	goodIdeas := m.ideasStorage.GetGoodIdeas()
	if len(*goodIdeas) > 0 {
		sb.WriteString("\n\n# Fav ideas\n")
		for _, v := range *goodIdeas {
			sb.WriteString(fmt.Sprintf("%s---", v))
		}
	}

	badIdeas := m.ideasStorage.GetBadIdeas()
	if len(*badIdeas) > 0 {
		sb.WriteString("\n\n# Ideas to be polished\n")
		for _, v := range *badIdeas {
			sb.WriteString(fmt.Sprintf("%s---", v))
		}
	}

	permissions := os.FileMode(0644)
	os.WriteFile(m.resultFilename, []byte(sb.String()), permissions)
	sb.Reset()
}

func (m *Model) updateRanking() {
	// Update good ideas
	goodIdeas := m.ideasStorage.GetGoodIdeas()
	goodIdeasContent, _ := glamour.Render(strings.Join(*goodIdeas, "---"), "dark")
	m.goodIdeasViewport.SetContent(goodIdeasContent)
	m.goodIdeasViewport.GotoBottom()

	// Update current
	rankingIdeaContent, _ := glamour.Render(m.ideasStorage.ideas[m.ideasStorage.currentRankingIdea], "dark")
	m.rankingIdeaViewport.SetContent(rankingIdeaContent)
	m.rankingIdeaViewport.GotoTop()

	// Update bad ideas
	badIdeas := m.ideasStorage.GetBadIdeas()
	badIdeasContent, _ := glamour.Render(strings.Join(*badIdeas, "---"), "dark")
	m.badIdeasViewport.SetContent(badIdeasContent)
	m.badIdeasViewport.GotoBottom()
}

func (m Model) View() string {
	// Header
	s := HeaderStyle.Render("//// Polygo /////////////////////////////")
	s += "\n\n"

	if m.timer.Running() {
		s += fmt.Sprintf("Time left: %s\n\n", m.timer.View())

		nextIdeaView := m.ideaInput.View()

		m.updateWrittenIdeasViewport()
		storedIdeasView := m.writtenIdeas.View()

		s += lipgloss.JoinHorizontal(lipgloss.Top, nextIdeaView, storedIdeasView)

		s += HelpStyle("\n\n[⇄] TAB - store an idea")
	}

	if m.focused == "minutes" || m.focused == "seconds" {
		s += fmt.Sprintf(
			"Set timer:\n%s: %s",
			m.minutesInput.View(),
			m.secondsInput.View(),
		)

		s += HelpStyle("\n\n[⇄] TAB - jump between minutes and seconds\n[⏎] ENTER - start a session\n[Ctrl+C] - exit")
	}

	if m.focused == "ranking" {
		s += "Time's up!\nLet's sort ideas ✨\nSwipe left for the fav ones, right for the nahs.\n\n"

		m.updateRanking()
		goodIdeasView := m.goodIdeasViewport.View()
		currentIdeaView := m.rankingIdeaViewport.View()
		badIdeasView := m.badIdeasViewport.View()

		s += lipgloss.JoinHorizontal(lipgloss.Top, goodIdeasView, currentIdeaView, badIdeasView)
		s += HelpStyle("\n\n[←] LEFT - like\n[→] RIGHT - need polishing\n[Ctrl+C] - exit")
	}

	if m.focused == "store" {
		confirmation, _ := glamour.Render(
			fmt.Sprintf("# DONE\n\nFile `%s` is saved!", m.resultFilename),
			"dark",
		)
		s += confirmation
		s += HelpStyle("\n\n[Ctrl+C] - exit")
	}

	return s
}
