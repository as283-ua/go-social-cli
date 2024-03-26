package mvc

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type LoginPage struct {
	username textinput.Model
	password textinput.Model
}

func InitialLoginModel() LoginPage {
	model := LoginPage{}

	model.username = textinput.New()
	model.username.Placeholder = "Username"
	model.username.Focus()

	model.password = textinput.New()
	model.password.Placeholder = "Password"

	return model
}

func (m LoginPage) Init() tea.Cmd {
	return nil
}

func (m LoginPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// peticion a servidor
			success := true
			if success {
				return InitialHomeModel(true), nil
			}
		}
	}
	return m, tea.Batch(passCmd, userCmd)
}

func (m LoginPage) View() string {
	var s string

	s = "Login\n\n"

	s += m.username.View() + "\n"
	s += m.password.View()

	s += "\n\nPresione 'q' para salir\n\n"

	return s
}
