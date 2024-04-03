package mvc

import (
	"bytes"
	"fmt"
	"io"
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
			return InitialHomeModel("", nil, m.client), nil
		case "ctrl+c":
			return m, tea.Quit
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

			return InitialHomeModel(m.username.Value(), token, m.client), nil
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
		s += "Info: "
		s += m.msg
		// charLimit := 100
		// for i := 0; i < len(m.msg)/charLimit; i++ {
		// 	if (i+1)*charLimit > len(m.msg) {
		// 		s += m.msg[i*charLimit:] + "\n"
		// 	}
		// 	s += m.msg[i*charLimit:(i+1)*charLimit] + "\n"
		// }
		s += "\n\n"
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

	if !r.Ok {
		return nil, fmt.Errorf(r.Msg)
	}

	token := r.Token

	return token, nil
}

func (m LoginPage) LoginCert() ([]byte, error) {
	resp, err := m.client.Get(fmt.Sprintf("https://localhost:10443/login/cert?user=%s", m.username.Value()))

	if err != nil {
		return nil, fmt.Errorf("error conectando con el servidor. %s", err.Error())
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("usuario no encontrado")
	}

	token := make([]byte, 32)
	io.ReadFull(resp.Body, token)
	resp.Body.Close()

	privateKey, err := util.ReadRSAKeyFromFile(fmt.Sprintf("%s.key", m.username.Value()))

	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo de clave. %s", err.Error())
	}

	signature, err := util.SignRSA(token, privateKey)

	if err != nil {
		return nil, fmt.Errorf("error firmando token para el servidor. %s", err.Error())
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://localhost:10443/login/cert?user=%s", m.username.Value()), bytes.NewReader(signature))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "text/plain")

	// pubkeybytes := util.ReadPublicKeyBytesFromFile(m.username.Value() + ".pub")
	// pubkey := util.ParsePublicKey(pubkeybytes)

	// err = util.CheckSignatureRSA(token, signature, pubkey)

	resp, err = m.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error conectando con el servidor. %s", err.Error())
	}

	r := model.Resp{}
	err = util.DecodeJSON(resp.Body, &r)

	if err != nil {
		return nil, fmt.Errorf("error decodificando JSON. %s", err.Error())
	}

	// return nil, fmt.Errorf("checks out %v, real token %v, signature %v, pubkey %v", err == nil, token, signature, pubkeybytes)
	// return nil, fmt.Errorf("checks out %v, real token %v, signature %v, pubkey %v", err == nil, len(token), len(signature), len(pubkeybytes))
	if !r.Ok {
		return nil, fmt.Errorf("%v", r)
	}

	// return nil, fmt.Errorf("%v", r)

	return util.Decode64(string(r.Token))
}
