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
	"testing"

	"github.com/charmbracelet/lipgloss"
	"gitub.com/atye/ttchat/internal/types"
)

// build a Private Message
// call callback function pasisng in pm
type irc struct {
	callback func(types.PrivateMessage)
}

// pass in the callback function
func (i *irc) OnPrivateMessage(f func(types.PrivateMessage)) {
	i.callback = f
}

func (i *irc) Say(string, string) {}
func TestGetMessageSource(t *testing.T) {
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
			incomingIRC := &irc{}
			i := NewIRCService(test.userDisplayName, "testChannel", incomingIRC)

			s := i.GetMessageSource()
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
			incomingIRC := &irc{}
			i := NewIRCService(test.name, "testChannel", incomingIRC)

			s := i.GetMessageSource()
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
