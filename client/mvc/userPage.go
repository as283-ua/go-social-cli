package mvc

import (
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
)

type UserPage struct {
	username string

	myUsername string
	client     *http.Client
	token      []byte
}

func InitialUserPageModel(myUsername string, token []byte, client *http.Client, username string) UserPage {
	model := UserPage{}
	model.client = client
	model.username = username
	model.myUsername = myUsername
	model.token = token
	return model
}

func (m UserPage) Init() tea.Cmd {
	return nil
}

func (m UserPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			return InitialHomeModel(m.myUsername, m.token, m.client), nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m UserPage) View() string {
	var s string
	return s
}
