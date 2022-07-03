package terminal

import (
	"fmt"
	"strings"

	"github.com/atye/ttchat/internal/types"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

type Twitch interface {
	GetMessageSource() <-chan types.Message
	Publish(string)
}

type Model struct {
	in          <-chan types.Message
	lines       []string
	ti          textinput.Model
	t           Twitch
	mode        mode
	height      int
	width       int
	lineSpacing int
}

type mode int

const (
	Initialize mode = iota
	Run
)

func NewModel(t Twitch, lineSpacing int) Model {
	ti := textinput.NewModel()
	ti.Placeholder = "Send a message"
	ti.Focus()

	return Model{
		in:          t.GetMessageSource(),
		mode:        Initialize,
		ti:          ti,
		t:           t,
		lineSpacing: lineSpacing,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		listenForMessages(m.in),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			return m, tea.Quit

		case tea.KeyCtrlU:
			m.ti.SetValue("")
			return m, listenForMessages(m.in)

		case tea.KeyEnter:
			if v := strings.TrimSpace(m.ti.Value()); v != "" {
				m.t.Publish(v)
				m.ti.SetValue("")
			}
			return m, listenForMessages(m.in)

		default:
			var cmd tea.Cmd
			m.ti, cmd = m.ti.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		switch m.mode {
		case Initialize:
			m.lines = make([]string, msg.Height-2)
			for i := 0; i < len(m.lines); i++ {
				m.lines[i] = "\n"
			}
			m.height = msg.Height
			m.width = msg.Width
			m.mode = Run
		}
		return m, listenForMessages(m.in)

	case types.Message:
		msgLines := strings.Split(wordwrap.String(fmt.Sprintf("%s: %s", msg.GetName(), msg.GetText()), m.width), "\n")
		for i := 0; i < len(msgLines); i++ {
			msgLines[i] = fmt.Sprintf("%s\n", msgLines[i])
		}
		m.lines = append(m.lines[len(msgLines):], msgLines...)

		for i := 0; i < m.lineSpacing; i++ {
			m.lines = append(m.lines[1:], "\n")
		}
		return m, listenForMessages(m.in)

	default:
		return m, listenForMessages(m.in)
	}
}

func (m Model) View() string {
	var b strings.Builder
	for i := 0; i < m.lineSpacing; i++ {
		b.WriteString("\n")
	}
	for i, line := range m.lines {
		if i >= len(m.lines)-m.lineSpacing {
			continue
		}
		b.WriteString(line)
	}

	b.WriteString("\n")
	b.WriteString(m.ti.View())
	return b.String()
}

func listenForMessages(in <-chan types.Message) tea.Cmd {
	return func() tea.Msg {
		return <-in
	}
}
