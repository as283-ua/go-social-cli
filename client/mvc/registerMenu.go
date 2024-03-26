package mvc

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"util"
	"util/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type RegisterPage struct {
	username textinput.Model
	password textinput.Model
	msg      string

	client *http.Client
}

func InitialRegisterModel(client *http.Client) RegisterPage {
	model := RegisterPage{}

	model.username = textinput.New()
	model.username.Placeholder = "Username"
	model.username.Focus()

	model.password = textinput.New()
	model.password.Placeholder = "Password"

	model.client = client

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
		case "left":
			return InitialHomeModel(false, nil, m.client), nil
		case "enter":
			token, err := m.Register()
			if err != nil {
				m.msg = err.Error()
				return m, nil
			}

			// token := []byte("token")
			return InitialHomeModel(true, token, m.client), nil
		}
	}
	return m, tea.Batch(passCmd, userCmd)
}

func (m RegisterPage) View() string {
	s := "Register\n\n"

	s += m.username.View() + "\n"
	s += m.password.View() + "\n\n"

	lines := 5

	if m.msg != "" {
		lines = 7
		s += "Info: " + m.msg + "\n\n"
	}

	_, y, _ := term.GetSize(0)

	for i := 0; i < y-lines; i++ {
		s += "\n"
	}

	return s
}

func (m RegisterPage) Register() ([]byte, error) {
	username := m.username.Value()
	password := m.password.Value()

	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password must not be empty")
	}

	var publicKeyBytes []byte
	var privateKey *rsa.PrivateKey
	if _, err := os.Stat(fmt.Sprintf("%s.key", username)); err != nil {
		// no hay err -> el archivo no existe
		pk, err := rsa.GenerateKey(rand.Reader, 2048)
		privateKey = pk
		util.FailOnError(err)

		// writeECDSAKeyToFile(fmt.Sprintf("%s.key", username), privateKey)
		util.WriteRSAKeyToFile(fmt.Sprintf("%s.key", username), privateKey)
		publicKeyBytes = util.WritePublicKeyToFile(fmt.Sprintf("%s.pub", username), &privateKey.PublicKey)
	} else {
		privateKey, err = util.ReadRSAKeyFromFile(fmt.Sprintf("%s.key", username))
		if err != nil {
			return nil, err
		}
		publicKeyBytes = util.ReadPublicKeyBytesFromFile(fmt.Sprintf("%s.pub", username))
	}

	register := model.RegisterCredentials{User: username, Pass: password, PubKey: publicKeyBytes}
	jsonBody := util.EncodeJSON(register)

	resp, err := m.client.Post("https://localhost:10443/register", "application/json", bytes.NewReader(jsonBody))

	if err != nil {
		return nil, fmt.Errorf("error al hacer la peticion. Servidor caído")
	}

	var r = model.Resp{}
	var token []byte
	util.DecodeJSON(resp.Body, &r)
	if !r.Ok {
		return nil, fmt.Errorf("%s, %s, %s", r.Msg, username, password)
	} else {
		msg, _ := util.Decode64(r.Msg)
		util.DecryptWithRSA(msg, privateKey)
		token = r.Token
	}

	resp.Body.Close()
	return token, nil
}
