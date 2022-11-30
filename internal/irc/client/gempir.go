package client

import (
	"fmt"
	"log"
	"time"

	"github.com/atye/ttchat/internal/irc"
	"github.com/atye/ttchat/internal/types"
	"github.com/gempir/go-twitch-irc/v2"
)

type Gempir struct {
	irc *twitch.Client
}

var _ irc.IRC = Gempir{}

func NewGempirClient(log *log.Logger, username string, channel string, accessToken string) Gempir {
	c := twitch.NewClient(username, fmt.Sprintf("oauth:%s", accessToken))
	c.Join(channel)
	go func() {
		for {
			err := c.Connect()
			if err != nil {
				log.Printf("disconnected from channel %s: %v\n", channel, err)
				log.Println("retrying...")
			}
			time.Sleep(5 * time.Second)
		}
	}()
	return Gempir{irc: c}
}

func (g Gempir) OnPrivateMessage(f func(types.PrivateMessage)) error {
	g.irc.OnPrivateMessage(func(message twitch.PrivateMessage) {
		f(types.PrivateMessage{
			Name:  message.User.DisplayName,
			Text:  message.Message,
			Color: message.User.Color,
		})
	})
	return nil
}

func (g Gempir) Publish(channel string, msg string) error {
	g.irc.Say(channel, msg)
	return nil
}
