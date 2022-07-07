package irc

import (
	"fmt"
	"strings"

	"github.com/atye/ttchat/internal/terminal"
	"github.com/atye/ttchat/internal/types"
	"github.com/charmbracelet/lipgloss"
)

type IRC interface {
	OnPrivateMessage(func(types.PrivateMessage))
	Say(string, string) // channel, message
}

type IRCService struct {
	displayName string
	channel     string
	irc         IRC
	upstream    chan types.Message
}

const (
	DefaultNameColor   = "#1E90FF" //Dodger Blue
	UserHighlightColor = "#6441A5" //Twitch purple
)

var (
	UserHighLightStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(UserHighlightColor))
)

var _ terminal.Twitch = IRCService{}

func NewIRCService(displayName string, channel string, irc IRC) IRCService {
	return IRCService{
		irc:         irc,
		displayName: displayName,
		channel:     channel,
		upstream:    make(chan types.Message),
	}
}

func (c IRCService) GetMessageSource() <-chan types.Message {
	c.irc.OnPrivateMessage(func(incoming types.PrivateMessage) {
		styled := incoming
		if styled.Color == "" {
			styled.Color = DefaultNameColor
		}

		styled.Text = highlightUserMentions(styled.Text, c.displayName)

		styled.Name = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(styled.Color)).Render(styled.Name)
		if incoming.Name == c.displayName {
			styled.Name = UserHighLightStyle.Render(c.displayName)
		}

		c.upstream <- styled
	})
	return c.upstream
}

func (c IRCService) Publish(msg string) {
	c.irc.Say(c.channel, msg)
	c.upstream <- types.PrivateMessage{
		Name: UserHighLightStyle.Render(c.displayName),
		Text: highlightUserMentions(msg, c.displayName),
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
