package mvc

import (
	"fmt"
	"net/http"
	"util/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HomePage struct {
	options     []string
	cursor      int
	cursorStyle lipgloss.Style

	client *http.Client
	user   model.User
}

func InitialHomeModel(user model.User, client *http.Client) HomePage {
	m := HomePage{}
	m.user = user

	m.client = client

	if m.user.Token == nil {
		m.options = []string{
			"Register",
			"Login",
			"Login with certificate",
			"Posts",
		}
	} else {
		m.options = []string{
			"Posts",
			"Search user",
			"Create group",
			"Join group",
			"See group posts",
			"Logout",
		}

		if m.user.Role == model.Admin {
			m.options = append([]string{"Block User"}, m.options...)
		}
	}

	m.cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#FFF"))

	return m
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
			if m.user.Token == nil {
				switch m.cursor {
				case 0:
					return InitialRegisterModel(m.client), nil
				case 1:
					return InitialLoginModel(m.client, false), nil
				case 2:
					return InitialLoginModel(m.client, true), nil
				case 3:
					cmd := GetPostsMsg(0, "", "", nil, m.client)
					m, _ := InitialPostListModel(model.User{}, "", m.client)
					return m, cmd
				}
			} else {
				switch m.cursor {
				case 0:
					cmd := GetPostsMsg(0, "", m.user.Name, m.user.Token, m.client)
					m, _ := InitialPostListModel(m.user, "", m.client)
					return m, cmd
				case 1:
					cmd := GetUserMsg(0, "", m.client)
					return InitialUserSearchPageModel(m.user, "", m.client), cmd
				case 2:
					return InitialAccessGroupModel(m.client, m.user, 1), nil
				case 3:
					return InitialAccessGroupModel(m.client, m.user, 2), nil
				case 4:
					return InitialAccessGroupModel(m.client, m.user, 3), nil
				case 5:
					return InitialHomeModel(model.User{}, m.client), nil
				}
			}
		}
	}
	return m, nil
}

func (m HomePage) View() string {
	var s string
	if m.user.Name != "" {
		s = fmt.Sprintf("Hola, %s\n\n", m.user.Name)
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
