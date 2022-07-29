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
	lines       []line
	ti          textinput.Model
	t           Twitch
	mode        mode
	width       int
	lineSpacing int
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
			m.lines = make([]line, msg.Height-2)
			for i := 0; i < len(m.lines); i++ {
				m.lines[i] = line{value: "\n"}
			}
			m.width = msg.Width
			m.mode = Run
		case Run:
			newLines := make([]line, msg.Height-2)
			newLinesIndex := len(newLines) - 1
			linesIndex := len(m.lines) - 1

			for i := 0; i < len(newLines); i++ {
				newLines[i] = line{value: "\n"}
			}

		out:
			for linesIndex >= 0 {
				if newLinesIndex < 0 {
					break
				}

				author := m.lines[linesIndex].author
				var buf []string
				if author != "" {
					for j := linesIndex; j >= 0; j-- {
						if m.lines[j].author == author {
							buf = append([]string{strings.Replace(m.lines[j].value, "\n", "", -1)}, buf...)
							linesIndex--
						} else {
							break
						}
					}
				} else {
					buf = []string{strings.Replace(m.lines[linesIndex].value, "\n", "", -1)}
					linesIndex--
				}

				msgLines := strings.Split(wordwrap.String(strings.Join(buf, " "), msg.Width), "\n")
				if len(msgLines) == 1 {
					newLines[newLinesIndex] = line{author: author, value: fmt.Sprintf("%s\n", msgLines[0])}
					newLinesIndex--
				} else {
					for j := len(msgLines) - 1; j >= 0; j-- {
						if newLinesIndex < 0 {
							break out
						}
						newLines[newLinesIndex] = line{author: author, value: fmt.Sprintf("%s\n", msgLines[j])}
						newLinesIndex--
					}
				}
			}
			m.lines = newLines
			m.width = msg.Width
		}
		return m, listenForMessages(m.in)

	case types.Message:
		for i := 0; i < m.lineSpacing; i++ {
			m.lines = append(m.lines[1:], line{value: "\n"})
		}

		msgLines := strings.Split(wordwrap.String(fmt.Sprintf("%s: %s", msg.GetName(), msg.GetText()), m.width), "\n")

		newLines := make([]line, len(msgLines))
		for i := 0; i < len(msgLines); i++ {
			newLines[i] = line{author: msg.GetName(), value: fmt.Sprintf("%s\n", msgLines[i])}
		}
		m.lines = append(m.lines[len(newLines):], newLines...)

		return m, listenForMessages(m.in)

	default:
		return m, listenForMessages(m.in)
	}
}

func (m Model) View() string {
	var b strings.Builder
	for _, line := range m.lines {
		b.WriteString(line.value)
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
