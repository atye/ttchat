package irc

import (
	"fmt"
	"log"
	"testing"

	"github.com/atye/ttchat/internal/types"
	"github.com/charmbracelet/lipgloss"
)

type mockIrc struct {
	callback func(types.PrivateMessage)
}

func (i *mockIrc) OnPrivateMessage(f func(types.PrivateMessage)) error {
	i.callback = f
	return nil
}

func (i *mockIrc) Publish(string, string) error { return nil }

func TestIncomingMessages(t *testing.T) {
	tests := []struct {
		Name            string
		userDisplayName string
		pm              types.PrivateMessage
		wantName        string
		wantText        string
	}{
		{
			"incoming default color",
			"",
			types.PrivateMessage{
				Name: "foo",
				Text: "bar",
			},
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(DefaultNameColor)).Render("foo"),
			"bar",
		},
		{
			"incoming with color",
			"",
			types.PrivateMessage{
				Name:  "foo",
				Text:  "bar",
				Color: "#000000",
			},
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#000000")).Render("foo"),
			"bar",
		},
		{
			"incoming mention",
			"user",
			types.PrivateMessage{
				Name: "foo",
				Text: "hi @user",
			},
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(DefaultNameColor)).Render("foo"),
			fmt.Sprintf("hi %s", UserHighLightStyle.Render("@user")),
		},
		{
			"incoming mention mix case",
			"User",
			types.PrivateMessage{
				Name: "foo",
				Text: "hi @user",
			},
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(DefaultNameColor)).Render("foo"),
			fmt.Sprintf("hi %s", UserHighLightStyle.Render("@user")),
		},
		{
			"incoming is you",
			"user",
			types.PrivateMessage{
				Name: "user",
				Text: "bar",
			},
			lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(UserHighlightColor)).Render("user"),
			"bar",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			incomingIRC := &mockIrc{}
			i := NewTwitch(incomingIRC, log.Default(), test.userDisplayName, "testChannel")

			s := i.IncomingMessages()
			go incomingIRC.callback(test.pm)

			m := <-s

			if m.GetName() != test.wantName {
				t.Errorf("expected name %s, got %s", test.wantName, m.GetName())
			}

			if m.GetText() != test.wantText {
				t.Errorf("expected name %s, got %s", test.wantText, m.GetText())
			}
		})
	}
}

func TestPublish(t *testing.T) {
	tests := []struct {
		Name     string
		name     string
		text     string
		wantName string
		wantText string
	}{
		{
			"publish message",
			"user",
			"testText",
			UserHighLightStyle.Render("user"),
			"testText",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			incomingIRC := &mockIrc{}
			i := NewTwitch(incomingIRC, log.Default(), test.name, "testChannel")

			s := i.IncomingMessages()
			go i.Publish(test.text)

			m := <-s

			if m.GetName() != test.wantName {
				t.Errorf("expected name %s, got %s", test.wantName, m.GetName())
			}

			if m.GetText() != test.wantText {
				t.Errorf("expected name %s, got %s", test.wantText, m.GetText())
			}
		})
	}
}
