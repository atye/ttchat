package terminal

import (
	"fmt"
	"strings"

	"github.com/atye/ttchat/internal/types"
	"github.com/muesli/reflow/wordwrap"
)

type IRC interface {
	IncomingMessages() <-chan types.Message
	Publish(string)
}

type Channel struct {
	name        string
	incomingMsg <-chan types.Message
	lines       []line
	irc         IRC
	width       int
	lineSpacing int
}

func NewChannel(irc IRC, name string, lineSpacing int) *Channel {
	return &Channel{
		name:        name,
		incomingMsg: irc.IncomingMessages(),
		irc:         irc,
		lineSpacing: lineSpacing,
	}
}

func (c *Channel) initLines(lines int) {
	c.lines = make([]line, lines)
	for i := 0; i < len(c.lines); i++ {
		c.lines[i] = line{value: "\n"}
	}
}

func (c *Channel) update(msg types.Message) {
	for i := 0; i < c.lineSpacing; i++ {
		c.lines = append(c.lines[1:], line{value: "\n"})
	}

	msgLines := strings.Split(wordwrap.String(fmt.Sprintf("%s: %s", msg.GetName(), msg.GetText()), c.width), "\n")

	newLines := make([]line, len(msgLines))
	for i := 0; i < len(msgLines); i++ {
		newLines[i] = line{author: msg.GetName(), value: fmt.Sprintf("%s\n", msgLines[i])}
	}
	c.lines = append(c.lines[len(newLines):], newLines...)
}

func (c *Channel) resize(height int, width int) {
	newLines := make([]line, height)
	newLinesIndex := len(newLines) - 1
	linesIndex := len(c.lines) - 1

	for i := 0; i < len(newLines); i++ {
		newLines[i] = line{value: "\n"}
	}

out:
	for linesIndex >= 0 {
		if newLinesIndex < 0 {
			break
		}

		author := c.lines[linesIndex].author
		var buf []string
		if author != "" {
			for j := linesIndex; j >= 0; j-- {
				if c.lines[j].author == author {
					buf = append([]string{strings.Replace(c.lines[j].value, "\n", "", -1)}, buf...)
					linesIndex--
				} else {
					break
				}
			}
		} else {
			buf = []string{strings.Replace(c.lines[linesIndex].value, "\n", "", -1)}
			linesIndex--
		}

		msgLines := strings.Split(wordwrap.String(strings.Join(buf, " "), width), "\n")
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
	c.lines = newLines
	c.width = width
}
