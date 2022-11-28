package terminal

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/atye/ttchat/internal/types"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	channels      []*Channel
	incomingMsg   <-chan types.Message
	log           *log.Logger
	activeChannel int
	tabs          string
	textInput     textinput.Model
	mode          mode
}

type line struct {
	value  string
	author string
}

type mode int

const (
	Initialize mode = iota
	Run
)

var (
	linesOffset = 5
)

func NewModel(log *log.Logger, channels ...*Channel) *Model {
	ti := textinput.NewModel()
	ti.Placeholder = "Send a message"
	ti.Focus()

	return &Model{
		channels:  channels,
		textInput: ti,
		mode:      Initialize,
		log:       log,
	}
}

func (m *Model) Init() tea.Cmd {
	incomingMsg := make(chan types.Message)
	for _, ch := range m.channels {
		ch := ch
		go func() {
			for {
				msg := <-ch.incomingMsg
				incomingMsg <- msg
			}
		}()
	}
	m.incomingMsg = incomingMsg
	m.activeChannel = 0
	m.setTabs(m.channels[m.activeChannel].name)

	return listenForMessages(m)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			return m, tea.Quit
		case tea.KeyCtrlU:
			m.textInput.SetValue("")
			return m, listenForMessages(m)
		case tea.KeyEnter:
			if v := strings.TrimSpace(m.textInput.Value()); v != "" {
				m.channels[m.activeChannel].irc.Publish(v)
				m.textInput.SetValue("")
			}
			return m, listenForMessages(m)
		case tea.KeyTab:
			if m.activeChannel+1 >= len(m.channels) {
				m.activeChannel = 0
			} else {
				m.activeChannel++
			}
			m.setTabs(m.channels[m.activeChannel].name)
		case tea.KeyShiftTab:
			if m.activeChannel-1 < 0 {
				m.activeChannel = len(m.channels) - 1
			} else {
				m.activeChannel--
			}
			m.setTabs(m.channels[m.activeChannel].name)
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		var wg sync.WaitGroup
		switch m.mode {
		case Initialize:
			for _, ch := range m.channels {
				ch := ch
				wg.Add(1)
				go func() {
					defer wg.Done()
					ch.initLines(msg.Height - linesOffset)
				}()
			}
			wg.Wait()
			m.mode = Run
		case Run:
			for _, ch := range m.channels {
				ch := ch
				wg.Add(1)
				go func() {
					defer wg.Done()
					ch.resize(msg.Height-linesOffset, msg.Width)
				}()
			}
			wg.Wait()
		}
		return m, listenForMessages(m)
	case types.Message:
		var ch *Channel
		for _, c := range m.channels {
			if c.name == msg.GetChannel() {
				ch = c
				break
			}
		}
		if ch == nil {
			m.log.Printf("no channel found for %s\n", msg.GetChannel())
			return m, listenForMessages(m)
		}
		ch.update(msg)
		return m, listenForMessages(m)

	default:
		return m, listenForMessages(m)
	}
	return m, listenForMessages(m)
}

func (m *Model) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\n", m.tabs))
	for _, line := range m.channels[m.activeChannel].lines {
		b.WriteString(line.value)
	}

	b.WriteString("\n")
	b.WriteString(m.textInput.View())
	return b.String()
}

var (
	highlight = lipgloss.AdaptiveColor{Light: "#efeff1", Dark: "#6441A5"}

	border = lipgloss.Border{
		Bottom: "─",
	}

	active    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6441A5")).Border(border).BorderForeground(highlight)
	nonActive = lipgloss.NewStyle().Border(border)
)

func (m *Model) setTabs(activeTabName string) {
	var tabs []string
	for _, ch := range m.channels {
		if ch.name == activeTabName {
			tabs = append(tabs, active.Render(ch.name))
		} else {
			tabs = append(tabs, nonActive.Render(ch.name))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	m.tabs = lipgloss.JoinHorizontal(lipgloss.Bottom, row)
}

func listenForMessages(m *Model) tea.Cmd {
	return func() tea.Msg {
		return <-m.incomingMsg
	}
}
