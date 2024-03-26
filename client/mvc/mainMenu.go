package mvc

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HomePage struct {
	options     []string
	cursor      int
	cursorStyle lipgloss.Style
}

func InitialHomeModel() HomePage {
	model := HomePage{}
	model.options = []string{
		"Register",
		"Login",
		"All posts",
	}
	model.cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#FFF"))
	return model
}

func (m HomePage) Init() tea.Cmd {
	return nil
}

func (m HomePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.cursor++
			if m.cursor >= len(m.options) {
				m.cursor = 0
			}
		case "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.options) - 1
			}
		case "q":
			return m, tea.Quit
		case "enter", "right":
			switch m.cursor {
			case 0:
				return InitialRegisterModel(), nil
			case 1:
				return InitialLoginModel(), nil
			}
		}
	}
	return m, nil
}

func (m HomePage) View() string {
	s := "Opciones disponibles:\n"

	for i, option := range m.options {
		if i == m.cursor {
			s += "\t" + m.cursorStyle.Render(option) + "\n"
		} else {
			s += "\t" + option + "\n"
		}
	}

	s += "\nPresione 'q' para salir\n\n"

	return s
}
