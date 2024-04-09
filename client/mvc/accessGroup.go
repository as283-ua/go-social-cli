package mvc

import (
	"net/http"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type AccessGroupPage struct {
	groupName textinput.Model
	msg       string

	client *http.Client
	action int

	username string
	token    []byte
}

func InitialAccessGroupModel(client *http.Client, username string, token []byte, action int) AccessGroupPage {
	model := AccessGroupPage{}

	model.groupName = textinput.New()
	model.groupName.Placeholder = "Group Name"
	model.groupName.Focus()

	model.client = client
	model.action = action

	model.username = username
	model.token = token

	return model
}

func (m AccessGroupPage) Init() tea.Cmd {
	return nil
}

func (m AccessGroupPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		passCmd tea.Cmd
		userCmd tea.Cmd
	)
	m.groupName, passCmd = m.groupName.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			return InitialHomeModel(m.username, m.token, m.client), nil
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			// var (
			// 	token []byte
			// 	err   error
			// )

			// if !m.cert {
			// 	token, err = m.Login()
			// } else {
			// 	token, err = m.LoginCert()
			// }

			// if err != nil {
			// 	m.msg = err.Error()
			// 	return m, nil
			// }

			// return InitialHomeModel(m.username.Value(), token, m.client), nil
		}
	}
	return m, tea.Batch(passCmd, userCmd)
}

func (m AccessGroupPage) View() string {
	var s string

	if m.action == 1 {
		s = "Create group\n\n"
	} else if m.action == 2 {
		s = "Join group\n\n"
	} else {
		s = "Access group\n\n"
	}

	s += m.groupName.View() + "\n"

	s += "\n"

	if m.msg != "" {
		s += "Info: "
		s += m.msg
		s += "\n\n"
	}

	return s
}
