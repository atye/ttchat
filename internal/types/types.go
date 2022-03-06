package types

type Message interface {
	GetName() string
	GetColor() string
	GetText() string
}

type PrivateMessage struct {
	Name  string
	Color string
	Text  string
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
