package mvc

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"util"
	"util/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type LoginPage struct {
	username textinput.Model
	password textinput.Model
	msg      string

	client *http.Client
	cert   bool
}

func InitialLoginModel(client *http.Client, cert bool) LoginPage {
	model := LoginPage{}

	model.username = textinput.New()
	model.username.Placeholder = "Username"
	model.username.Focus()

	model.password = textinput.New()
	model.password.Placeholder = "Password"

	model.client = client
	model.cert = cert

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
			if !m.cert {
				m.password.Focus()
				m.username.Blur()
			}
		case "up":
			if !m.cert {
				m.username.Focus()
				m.password.Blur()
			}
		case "left":
			return InitialHomeModel(false, nil, m.client), nil
		case "enter":
			var (
				token []byte
				err   error
			)

			if !m.cert {
				token, err = m.Login()
			} else {
				token, err = m.LoginCert()
			}

			if err != nil {
				m.msg = err.Error()
				return m, nil
			}

			return InitialHomeModel(true, token, m.client), nil
		}
	}
	return m, tea.Batch(passCmd, userCmd)
}

func (m LoginPage) View() string {
	var s string

	s = "Login\n\n"

	s += m.username.View() + "\n"
	if !m.cert {
		s += m.password.View() + "\n"
	}

	s += "\n"

	if m.msg != "" {
		s += "Info: " + m.msg + "\n\n"
	}

	return s
}

func (m LoginPage) Login() ([]byte, error) {
	username := m.username.Value()
	password := m.password.Value()

	register := model.Credentials{User: strings.TrimSpace(username), Pass: strings.TrimSpace(password)}
	jsonBody := util.EncodeJSON(register)

	resp, err := m.client.Post("https://localhost:10443/login", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error al hacer la peticion")
	}

	var r = model.Resp{}
	util.DecodeJSON(resp.Body, &r)
	defer resp.Body.Close()
	r.Msg = string(util.Decode64(r.Msg))

	if !r.Ok {
		return nil, fmt.Errorf("credenciales invalidas")
	}

	token := r.Token

	return token, nil
}

func (m LoginPage) LoginCert() ([]byte, error) {
	privateKey, err := util.ReadRSAKeyFromFile(fmt.Sprintf("%s.key", m.username.Value()))

	if err != nil {
		return nil, err
	}

	return nil, nil
}
