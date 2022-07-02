package terminal

import (
	"fmt"
	"log"
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
	in          <-chan types.Message
	messages    []types.Message
	ti          textinput.Model
	t           Twitch
	mode        mode
	height      int
	lineSpacing int
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
			numMessages := (msg.Height - 2) / (m.lineSpacing + 1)
			msgs := make([]types.Message, numMessages)
			for i := 0; i < numMessages; i++ {
				msgs[i] = types.Message(noOpMessage{})
			}

			m.height = msg.Height
			m.messages = msgs
			m.mode = Run
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

	totalLines := len(m.messages)*(m.lineSpacing+1) + 2
	log.Printf("lines: %v, height: %v", totalLines, m.height)
	if totalLines < m.height {
		for i := totalLines; i <= m.height; i++ {
			b.WriteString("\n")
		}
	}

	for i := 0; i < len(m.messages); i++ {
		msg := m.messages[i]
		if msg.GetName() == "" || msg.GetText() == "" {
			b.WriteString("\n")

			if i != len(m.messages)-1 {
				for i := 0; i < m.lineSpacing; i++ {
					b.WriteString("\n")
				}
			}
			continue
		}

		b.WriteString(fmt.Sprintf("%s: %s\n", msg.GetName(), msg.GetText()))

		if i != len(m.messages)-1 {
			for i := 0; i < m.lineSpacing; i++ {
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
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
