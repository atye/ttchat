package terminal

import (
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/atye/ttchat/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

type mockIrc struct {
	msg              <-chan types.Message
	publishCallCount int
}

func (m *mockIrc) IncomingMessages() <-chan types.Message {
	return m.msg
}

func (m *mockIrc) Publish(string) {
	m.publishCallCount++
}

func TestModel(t *testing.T) {
	irc := &mockIrc{msg: make(chan types.Message)}

	one := NewChannel(irc, "one", 1)
	two := NewChannel(irc, "two", 1)

	sut := NewModel(log.New(io.Discard, "", log.LstdFlags), one, two)

	sut.Init()
	sut.Update(tea.WindowSizeMsg{
		Height: 15,
		Width:  20,
	})

	t.Run("Update", func(t *testing.T) {
		tests := []struct {
			name   string
			msg    []tea.Msg
			testFn func(*testing.T)
		}{
			{
				"new messages that fit in one line",
				[]tea.Msg{
					types.PrivateMessage{
						Channel: "one",
						Name:    "anon",
						Color:   "",
						Text:    "Pog",
					},
					types.PrivateMessage{
						Channel: "two",
						Name:    "anon",
						Color:   "",
						Text:    "Pog",
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0: "\n",
						1: "\n",
						2: "\n",
						3: "\n",
						4: "\n",
						5: "\n",
						6: "\n",
						7: "\n",
						8: "\n",
						9: "anon: Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n", want[i], line.value)
							}
						}
					}
				},
			},
			{
				"new messages are wrapped",
				[]tea.Msg{
					types.PrivateMessage{
						Channel: "one",
						Name:    "anon",
						Color:   "",
						Text:    "Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog",
					},
					types.PrivateMessage{
						Channel: "two",
						Name:    "anon",
						Color:   "",
						Text:    "Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog",
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0: "\n",
						1: "\n",
						2: "\n",
						3: "\n",
						4: "\n",
						5: "anon: Pog\n",
						6: "\n",
						7: "anon: Pog Pog Pog\n",
						8: "Pog Pog Pog Pog Pog\n",
						9: "Pog Pog Pog Pog Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n", want[i], line.value)
							}
						}
					}
				},
			},
			{
				"new messages from new author are wrapped",
				[]tea.Msg{
					types.PrivateMessage{
						Channel: "one",
						Name:    "anon2",
						Color:   "",
						Text:    "Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog",
					},
					types.PrivateMessage{
						Channel: "two",
						Name:    "anon2",
						Color:   "",
						Text:    "Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog",
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0: "\n",
						1: "anon: Pog\n",
						2: "\n",
						3: "anon: Pog Pog Pog\n",
						4: "Pog Pog Pog Pog Pog\n",
						5: "Pog Pog Pog Pog Pog\n",
						6: "\n",
						7: "anon2: Pog Pog Pog\n",
						8: "Pog Pog Pog Pog Pog\n",
						9: "Pog Pog Pog Pog Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n at index %d", want[i], line.value, i)
							}
						}
					}
				},
			},
			{
				"window width is increased by a lot",
				[]tea.Msg{
					tea.WindowSizeMsg{
						Height: 15,
						Width:  100,
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0: "\n",
						1: "\n",
						2: "\n",
						3: "\n",
						4: "\n",
						5: "anon: Pog\n",
						6: "\n",
						7: "anon: Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog\n",
						8: "\n",
						9: "anon2: Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n", want[i], line.value)
							}
						}
					}
				},
			},
			{
				"window width is decreased to interrupt message",
				[]tea.Msg{
					tea.WindowSizeMsg{
						Height: 15,
						Width:  14,
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0: "Pog Pog Pog\n",
						1: "Pog Pog Pog\n",
						2: "Pog Pog Pog\n",
						3: "Pog Pog\n",
						4: "\n",
						5: "anon2: Pog Pog\n",
						6: "Pog Pog Pog\n",
						7: "Pog Pog Pog\n",
						8: "Pog Pog Pog\n",
						9: "Pog Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n", want[i], line.value)
							}
						}
					}
				},
			},
			{
				"window width is reset to default",
				[]tea.Msg{
					tea.WindowSizeMsg{
						Height: 15,
						Width:  20,
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0: "\n",
						1: "\n",
						2: "\n",
						3: "Pog Pog Pog Pog Pog\n",
						4: "Pog Pog Pog Pog Pog\n",
						5: "Pog\n",
						6: "\n",
						7: "anon2: Pog Pog Pog\n",
						8: "Pog Pog Pog Pog Pog\n",
						9: "Pog Pog Pog Pog Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n", want[i], line.value)
							}
						}
					}
				},
			},
			{
				"window height is increased",
				[]tea.Msg{
					tea.WindowSizeMsg{
						Height: 17,
						Width:  20,
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0:  "\n",
						1:  "\n",
						2:  "\n",
						3:  "\n",
						4:  "\n",
						5:  "Pog Pog Pog Pog Pog\n",
						6:  "Pog Pog Pog Pog Pog\n",
						7:  "Pog\n",
						8:  "\n",
						9:  "anon2: Pog Pog Pog\n",
						10: "Pog Pog Pog Pog Pog\n",
						11: "Pog Pog Pog Pog Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n", want[i], line.value)
							}
						}
					}
				},
			},
			{
				"window height is decreased",
				[]tea.Msg{
					tea.WindowSizeMsg{
						Height: 15,
						Width:  20,
					},
				},
				func(t *testing.T) {
					want := map[int]string{
						0: "\n",
						1: "\n",
						2: "\n",
						3: "Pog Pog Pog Pog Pog\n",
						4: "Pog Pog Pog Pog Pog\n",
						5: "Pog\n",
						6: "\n",
						7: "anon2: Pog Pog Pog\n",
						8: "Pog Pog Pog Pog Pog\n",
						9: "Pog Pog Pog Pog Pog\n",
					}
					for _, ch := range sut.channels {
						for i, line := range ch.lines {
							if line.value != want[i] {
								t.Errorf("want %s\n, got %s\n", want[i], line.value)
							}
						}
					}
				},
			},
			{
				"message is typed",
				[]tea.Msg{
					tea.KeyMsg(tea.Key{
						Type:  tea.KeyRunes,
						Runes: []rune{'a'},
					}),
				},
				func(t *testing.T) {
					want := "a"
					if want != sut.textInput.Value() {
						t.Errorf("want %s, got %s", want, sut.textInput.Value())
					}
				},
			},
			{
				"message is sent",
				[]tea.Msg{
					tea.KeyMsg(tea.Key{
						Type: tea.KeyEnter,
					}),
				},
				func(t *testing.T) {
					if sut.textInput.Value() != "" {
						t.Errorf("expected text input to be empty, got %s", sut.textInput.Value())
					}
					if irc.publishCallCount != 1 {
						t.Errorf("expected 1 irc publish call")
					}
				},
			},
			{
				"tab is pressed twice",
				[]tea.Msg{
					tea.KeyMsg(tea.Key{
						Type: tea.KeyTab,
					}),
					tea.KeyMsg(tea.Key{
						Type: tea.KeyTab,
					}),
				},
				func(t *testing.T) {
					if sut.activeChannel != 0 {
						t.Errorf("expected active channel 0, got %d", sut.activeChannel)
					}
				},
			},
			{
				"shift tab is pressed twice",
				[]tea.Msg{
					tea.KeyMsg(tea.Key{
						Type: tea.KeyShiftTab,
					}),
					tea.KeyMsg(tea.Key{
						Type: tea.KeyShiftTab,
					}),
				},
				func(t *testing.T) {
					if sut.activeChannel != 0 {
						t.Errorf("expected active channel 0, got %d", sut.activeChannel)
					}
				},
			},
			{
				"clear text input",
				[]tea.Msg{
					tea.KeyMsg(tea.Key{
						Type:  tea.KeyRunes,
						Runes: []rune{'a'},
					}),
					tea.KeyMsg(tea.Key{
						Type: tea.KeyCtrlU,
					}),
				},
				func(t *testing.T) {
					if got := sut.textInput.Value(); got != "" {
						t.Errorf("want empty text input, got %s", got)
					}
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				for _, msg := range tc.msg {
					sut.Update(msg)
				}
				tc.testFn(t)
			})
		}

		_ = sut.View()
	})
}

func prettyPrintLines(v []line) {
	for i, line := range v {
		fmt.Printf("%d: %s", i, line.value)
	}
}
