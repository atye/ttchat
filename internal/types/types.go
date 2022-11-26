package types

type Message interface {
	GetChannel() string
	GetName() string
	GetColor() string
	GetText() string
}

type PrivateMessage struct {
	Channel string
	Name    string
	Color   string
	Text    string
}

func (m PrivateMessage) GetChannel() string {
	return m.Channel
}

func (m PrivateMessage) GetName() string {
	return m.Name
}

func (m PrivateMessage) GetText() string {
	return m.Text
}

func (m PrivateMessage) GetColor() string {
	return m.Color
}
