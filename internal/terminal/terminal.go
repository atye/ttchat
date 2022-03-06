package terminal

import (
	"fmt"
	"strings"

	"github.com/atye/ttchat/internal/types"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Twitch interface {
	GetMessageSource() <-chan types.Message
	Publish(string)
}

type Model struct {
	in       <-chan types.Message
	messages []types.Message
	ti       textinput.Model
	t        Twitch
	mode     mode
}

type mode int

const (
	Initialize mode = iota
	Run
)

type noOpMessage struct{}

func (m noOpMessage) GetName() string {
	return ""
}

func (m noOpMessage) GetText() string {
	return ""
}

func (m noOpMessage) GetColor() string {
	return ""
}

func (m noOpMessage) FromMyself() bool {
	return false
}

func NewModel(t Twitch) Model {
	ti := textinput.NewModel()
	ti.Placeholder = "Send a message"
	ti.Focus()

	return Model{
		in:   t.GetMessageSource(),
		mode: Initialize,
		ti:   ti,
		t:    t,
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
			msgs := make([]types.Message, msg.Height/2)
			for i := 0; i < msg.Height/2; i++ {
				msgs[i] = types.Message(noOpMessage{})
			}
			m.messages = msgs
			m.mode = Run
		default:
		}
		return m, listenForMessages(m.in)

	case types.Message:
		m.messages = append(m.messages[1:], msg)
		return m, listenForMessages(m.in)

	default:
		return m, listenForMessages(m.in)
	}
}

func (m Model) View() string {
	var b strings.Builder
	for _, msg := range m.messages {
		if msg.GetName() == "" || msg.GetText() == "" {
			b.WriteString("\n\n")
			continue
		}
		fmt.Fprintf(&b, "%s: %s\n\n", msg.GetName(), msg.GetText())
	}

	_, err := b.WriteString(m.ti.View())
	if err != nil {
		return ""
	}
	return b.String()
}

func listenForMessages(in <-chan types.Message) tea.Cmd {
	return func() tea.Msg {
		return <-in
	}
}
