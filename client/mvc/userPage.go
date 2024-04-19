package mvc

import (
	"net/http"
	"util/model"

	tea "github.com/charmbracelet/bubbletea"
)

type UserPage struct {
	username string

	user   model.User
	client *http.Client
}

func InitialUserPageModel(user model.User, client *http.Client, username string) UserPage {
	model := UserPage{}
	model.client = client
	model.username = username
	model.user = user
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
			return InitialHomeModel(m.user, m.client), nil
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
