package mvc

import (
	"fmt"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HomePage struct {
	options     []string
	cursor      int
	cursorStyle lipgloss.Style
	username    string

	client *http.Client
	token  []byte
}

func InitialHomeModel(username string, token []byte, client *http.Client) HomePage {
	model := HomePage{}
	model.username = username

	model.client = client
	model.token = token

	if model.token == nil {
		model.options = []string{
			"Register",
			"Login",
			"Login with certificate",
			"Posts",
		}
	} else {
		model.options = []string{
			"Posts",
			"Search user",
			"Create group",
			"Join group",
			"Groups",
			"Logout",
		}
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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", "right":
			if m.token == nil {
				switch m.cursor {
				case 0:
					return InitialRegisterModel(m.client), nil
				case 1:
					return InitialLoginModel(m.client, false), nil
				case 2:
					return InitialLoginModel(m.client, true), nil
				case 3:
					cmd := GetPostsMsg(0, m.client)
					return InitialPostListModel("", nil, m.client), cmd
				}
			} else {
				switch m.cursor {
				case 0:
					cmd := GetPostsMsg(0, m.client)
					return InitialPostListModel(m.username, m.token, m.client), cmd
				case 1:
					cmd := GetUserMsg(0, "", m.client)
					return InitialUserSearchPageModel(m.username, m.token, "", m.client), cmd
				case 2:
					return InitialAccessGroupModel(m.client, m.username, m.token, 1), nil
				case 3:
					return InitialAccessGroupModel(m.client, m.username, m.token, 2), nil
				case 4:
					return InitialAccessGroupModel(m.client, m.username, m.token, 3), nil
				case 5:
					return InitialHomeModel("", nil, m.client), nil
				}
			}
		}
	}
	return m, nil
}

func (m HomePage) View() string {
	var s string
	if m.username != "" {
		s = fmt.Sprintf("Hola, %s\n\n", m.username)
	}
	s += "Opciones disponibles:\n"

	for i, option := range m.options {
		if i == m.cursor {
			s += "\t" + m.cursorStyle.Render(option) + "\n"
		} else {
			s += "\t" + option + "\n"
		}
	}

	s += "\nPresione 'q' o 'ctrl-c' para salir\n\n"

	return s
}
