package irc

import (
	"fmt"
	"log"
	"strings"

	"github.com/atye/ttchat/internal/terminal"
	"github.com/atye/ttchat/internal/types"
	"github.com/charmbracelet/lipgloss"
)

// Generic interface for doing something with an IRC connection
type IRC interface {
	OnPrivateMessage(func(types.PrivateMessage)) error
	Publish(string, string) error // channel, message
}

type Twitch struct {
	displayName string
	channel     string
	irc         IRC
	upstream    chan types.Message
	log         *log.Logger
}

const (
	DefaultNameColor   = "#1E90FF" //Dodger Blue
	UserHighlightColor = "#6441A5" //Twitch purple
)

var (
	UserHighLightStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(UserHighlightColor))
)

var _ terminal.IRC = Twitch{}

func NewTwitch(irc IRC, log *log.Logger, displayName string, channel string) Twitch {
	s := Twitch{
		irc:         irc,
		displayName: displayName,
		channel:     channel,
		upstream:    make(chan types.Message),
		log:         log,
	}

	err := s.irc.OnPrivateMessage(func(incoming types.PrivateMessage) {
		styled := incoming
		styled.Channel = channel
		if styled.Color == "" {
			styled.Color = DefaultNameColor
		}

		styled.Text = highlightUserMentions(styled.Text, s.displayName)

		styled.Name = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(styled.Color)).Render(styled.Name)
		if incoming.Name == s.displayName {
			styled.Name = UserHighLightStyle.Render(s.displayName)
		}

		s.upstream <- styled
	})
	defer func() {
		if err != nil {
			s.log.Printf("irc: setting OnPrivateMessage behavior: %v\n", err)
		}
	}()

	return s
}

func (c Twitch) IncomingMessages() <-chan types.Message {
	return c.upstream
}

func (c Twitch) Publish(msg string) {
	c.irc.Publish(c.channel, msg)
	c.upstream <- types.PrivateMessage{
		Name:    UserHighLightStyle.Render(c.displayName),
		Text:    highlightUserMentions(msg, c.displayName),
		Channel: c.channel,
	}
}

func highlightUserMentions(text string, displayName string) string {
	texts := strings.Split(text, " ")
	for i, w := range texts {
		if strings.Contains(strings.ToLower(w), fmt.Sprintf("@%s", strings.ToLower(displayName))) {
			texts[i] = UserHighLightStyle.Render(w)
		}
	}
	return strings.Join(texts, " ")
}
