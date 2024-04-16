package mvc

import (
	"bytes"
	"net/http"
	"strings"
	"util"
	"util/model"

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
			switch m.action {
			case 1:
				if strings.TrimSpace(m.groupName.Value()) != "" {
					createGroup(&m)
				} else {
					m.msg = "Debes introducir una cadena"
				}
			case 2:
				if strings.TrimSpace(m.groupName.Value()) != "" {
					joinGroup(&m)
				} else {
					m.msg = "Debes introducir una cadena"
				}
			case 3:
				//andres things
			}
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

func createGroup(m *AccessGroupPage) {
	post := model.Group{Name: strings.TrimSpace(m.groupName.Value())}
	jsonBody := util.EncodeJSON(post)

	req, err := http.NewRequest("POST", "https://localhost:10443/groups", bytes.NewReader(jsonBody))
	util.FailOnError(err)
	req.Header.Add("content-type", "application/json")

	req.Header.Add("Authorization", util.Encode64(m.token))
	req.Header.Add("Username", m.username)

	resp, err := m.client.Do(req)
	if err != nil {
		m.msg = err.Error()
	}

	var r model.Resp
	util.DecodeJSON(resp.Body, &r)
	if r.Ok {
		m.msg = "Grupo creado correctamente"
	} else {
		m.msg = r.Msg
	}
}

func joinGroup(m *AccessGroupPage) {
	req, err := http.NewRequest("POST", "https://localhost:10443/groups/"+m.groupName.Value(), nil)
	util.FailOnError(err)

	req.Header.Add("Authorization", util.Encode64(m.token))
	req.Header.Add("Username", m.username)

	resp, err := m.client.Do(req)
	if err != nil {
		m.msg = err.Error()
	}

	var r model.Resp
	util.DecodeJSON(resp.Body, &r)
	if r.Ok {
		m.msg = "Ahora eres miembro del grupo " + m.groupName.Value()
	} else {
		m.msg = r.Msg
	}
}
