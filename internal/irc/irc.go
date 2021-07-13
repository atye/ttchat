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

package irc

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gitub.com/atye/ttchat/internal/terminal"
	"gitub.com/atye/ttchat/internal/types"
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
			styled.Name = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(UserHighlightColor)).Render(c.displayName)
		}

		c.upstream <- styled
	})
	return c.upstream
}

func (c IRCService) Publish(msg string) {
	c.irc.Say(c.channel, msg)

	styled := highlightUserMentions(msg, c.displayName)
	c.upstream <- types.PrivateMessage{
		Name: lipgloss.NewStyle().Background(lipgloss.Color(UserHighlightColor)).Render(c.displayName),
		Text: styled,
	}
}

func highlightUserMentions(text string, displayName string) string {
	texts := strings.Split(text, " ")
	for i, w := range texts {
		if strings.Contains(w, fmt.Sprintf("@%s", displayName)) {
			texts[i] = lipgloss.NewStyle().Background(lipgloss.Color(UserHighlightColor)).Render(w)
		}
	}
	return strings.Join(texts, " ")
}
