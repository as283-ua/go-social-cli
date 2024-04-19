package mvc

import (
	"bytes"
	"fmt"
	"net/http"
	"util"
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
				err := m.requestBlock()

				if err != nil {
					m.msg = err.Error()
				} else {
					var accion string
					if m.block {
						accion = "Bloqueado"
					} else {
						accion = "Desbloqueado"
					}
					m.msg = fmt.Sprintf("%v usuario %v", accion, m.username.Value())
				}
			}
		}
	}
	return m, tea.Batch(userCmd)
}

func (m BlockUserPage) View() string {
	s := "Block/unblock user\n\n"

	s += m.username.View() + "\n"

	if m.cursor == 1 {
		s += m.selectStyle.Render(">") + " "
	} else {
		s += "> "
	}

	if m.block {
		s += m.selectStyle.Render(m.blockOptions[0]) + "/" + m.blockOptions[1]
	} else {
		s += m.blockOptions[0] + "/" + m.selectStyle.Render(m.blockOptions[1])
	}

	if m.selectingBlockOption {
		s += " " + m.selectStyle.Render("<") + "\n"
	} else {
		s += "\n"
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

func (m BlockUserPage) requestBlock() error {
	block := model.Block{Blocked: m.block}

	blockJson := util.EncodeJSON(block)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://127.0.0.1:10443/users/%v/block", m.username.Value()), bytes.NewReader(blockJson))

	req.Header.Add("Authorization", util.Encode64(m.user.Token))
	req.Header.Add("Username", m.user.Name)

	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error actualizando estado del usuario %v. Status code: %v", m.username.Value(), resp.Status)
	}

	return nil
}
