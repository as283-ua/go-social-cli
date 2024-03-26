package mvc

import (
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HomePage struct {
	options     []string
	cursor      int
	cursorStyle lipgloss.Style
	loggedIn    bool

	client    *http.Client
	userToken []byte
}

func InitialHomeModel(loggedIn bool, token []byte, client *http.Client) HomePage {
	model := HomePage{}
	model.loggedIn = loggedIn

	if !loggedIn {
		model.options = []string{
			"Register",
			"Login",
			"Login with certificate",
			"All posts",
		}
	} else {
		model.options = []string{
			"All posts",
			"Post",
			"Search user",
			"Logout",
		}
	}

	model.cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#FFF"))

	model.client = client
	model.userToken = token

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
			if !m.loggedIn {
				switch m.cursor {
				case 0:
					return InitialRegisterModel(m.client), nil
				case 1:
					return InitialLoginModel(m.client, false), nil
				case 2:
					return InitialLoginModel(m.client, true), nil
				case 3:
					// get all posts
				}
			} else {
				switch m.cursor {
				case 3:
					return InitialHomeModel(false, nil, m.client), nil
				}
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
