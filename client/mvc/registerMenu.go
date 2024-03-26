package mvc

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type RegisterPage struct {
	username textinput.Model
	password textinput.Model
}

func InitialRegisterModel() RegisterPage {
	model := RegisterPage{}

	model.username = textinput.New()
	model.username.Placeholder = "Username"
	model.username.Focus()

	model.password = textinput.New()
	model.password.Placeholder = "Password"

	return model
}

func (m RegisterPage) Init() tea.Cmd {
	return nil
}

func (m RegisterPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		passCmd tea.Cmd
		userCmd tea.Cmd
	)
	m.password, passCmd = m.password.Update(msg)
	m.username, userCmd = m.username.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.password.Focus()
			m.username.Blur()
		case "up":
			m.username.Focus()
			m.password.Blur()
		case "q":
			return m, tea.Quit
		case "left":
			return InitialHomeModel(false), nil
		case "enter":
			success := true
			if success {
				return InitialHomeModel(true), nil
			}
		}
	}
	return m, tea.Batch(passCmd, userCmd)
}

func (m RegisterPage) View() string {
	var s string

	s = "Register\n\n"

	s += m.username.View() + "\n"
	s += m.password.View()

	s += "\n\nPresione 'q' para salir\n\n"

	return s
}
