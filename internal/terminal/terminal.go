// Copyright © 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gitub.com/atye/ttchat/internal/types"
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
}

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

func NewModel(height int, t Twitch) Model {
	ti := textinput.NewModel()
	ti.Placeholder = "Send a message"
	ti.Focus()

	var m []types.Message
	for i := 0; i < height; i++ {
		m = append(m, types.Message(noOpMessage{}))
	}

	return Model{
		in:       t.GetMessageSource(),
		messages: m,
		ti:       ti,
		t:        t,
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
