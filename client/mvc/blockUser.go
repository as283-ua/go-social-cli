package mvc

import (
	"net/http"
	"util/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BlockUserPage struct {
	username     textinput.Model
	block        bool
	blockOptions []string
	msg          string

	selectingBlockOption bool
	selectStyle          lipgloss.Style
	cursor               int

	client *http.Client
	user   model.User
}

func InitialBlockUserModel(user model.User, client *http.Client) BlockUserPage {
	m := BlockUserPage{}

	m.username = textinput.New()
	m.username.Placeholder = "Username"
	m.username.Focus()

	m.client = client
	m.user = user

	m.block = true
	m.blockOptions = []string{"Block", "Unblock"}

	m.selectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#FFF"))

	return m
}

func (m BlockUserPage) Init() tea.Cmd {
	return nil
}

func (m BlockUserPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		userCmd tea.Cmd
	)
	m.username, userCmd = m.username.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			if m.selectingBlockOption {
				m.block = !m.block
			} else {
				m.cursor++
			}
		case "up":
			if m.selectingBlockOption {
				m.block = !m.block
			} else {
				m.cursor--
			}
		}

		if m.cursor < 0 {
			m.cursor = 2
		} else if m.cursor >= 3 {
			m.cursor = 0
		}

		switch m.cursor {
		case 0:
			m.username.Focus()
		case 1:
			m.username.Blur()
		case 2:
			m.username.Blur()
		default:
			m.cursor = 0
		}

		switch msg.String() {
		case "left":
			return InitialHomeModel(m.user, m.client), nil
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			switch m.cursor {
			case 1:
				m.selectingBlockOption = !m.selectingBlockOption
			case 2:
				m.requestBlock()
			}
		}
	}
	return m, tea.Batch(userCmd)
}

func (m BlockUserPage) View() string {
	s := "Block/unblock user\n\n"

	s += m.username.View() + "\n"

	if m.cursor == 1 {
		s += m.selectStyle.Render("Block/Unblock") + "\n"
	}

	if m.selectingBlockOption {
		if m.block {
			s += "\t" + m.selectStyle.Render(m.blockOptions[0]) + "\n"
			s += "\t" + m.blockOptions[1] + "\n"
		} else {
			s += "\t" + m.blockOptions[0] + "\n"
			s += "\t" + m.selectStyle.Render(m.blockOptions[1]) + "\n"
		}
	} else {
		s += "\n\n"
	}

	if m.cursor == 2 {
		s += m.selectStyle.Render("[Submit]") + "\n"
	} else {
		s += "[Submit]" + "\n"
	}

	if m.msg != "" {
		s += "Info: " + m.msg + "\n\n"
	}

	return s
}

func (m BlockUserPage) requestBlock() {

}
