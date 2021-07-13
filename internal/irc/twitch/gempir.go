package twitch

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

import (
	"fmt"

	"github.com/atye/ttchat/internal/irc"
	"github.com/atye/ttchat/internal/types"
	"github.com/gempir/go-twitch-irc/v2"
)

type Gempir struct {
	irc *twitch.Client
}

var _ irc.IRC = Gempir{}

func NewGempirClient(username string, channel string, accessToken string) Gempir {
	c := twitch.NewClient(username, fmt.Sprintf("oauth:%s", accessToken))
	c.Join(channel)
	go func() {
		c.Connect()
	}()

	return Gempir{irc: c}
}

func (g Gempir) OnPrivateMessage(f func(types.PrivateMessage)) {
	g.irc.OnPrivateMessage(func(message twitch.PrivateMessage) {
		f(types.PrivateMessage{
			Name:  message.User.DisplayName,
			Text:  message.Message,
			Color: message.User.Color,
		})
	})
}

func (g Gempir) Say(channel string, msg string) {
	g.irc.Say(channel, msg)
}
