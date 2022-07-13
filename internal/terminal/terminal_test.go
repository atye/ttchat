package terminal

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate(t *testing.T) {
	t.Run("WindowSizeMsg", func(t *testing.T) {
		model := Model{
			mode: Run,
			lines: []line{
				{
					author: "test",
					value:  "that was crazy wow",
				},
				{
					author: "test1",
					value:  "that was crazy",
				},
			},
		}

		modelI, _ := model.Update(tea.WindowSizeMsg{Height: 10, Width: 8})
		m := modelI.(Model)
		fmt.Println(m.lines)

	})
}
