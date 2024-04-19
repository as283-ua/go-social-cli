package mvc

import (
	"client/message"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
	"util/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	usersPerReq = 5
	listSize    = 5
)

type UsernamesMsg []string

type UserSearchPage struct {
	usernames    []string
	searchBar    textinput.Model
	selectedUser int
	onSearchBtn  bool
	msg          string

	cursorStyle lipgloss.Style

	searched        string
	myUsername      string
	client          *http.Client
	token           []byte
	pagesLoaded     int
	loadedUsernames map[string]bool
	canRequestMore  bool
}

/*
usernames es la lista de usuarios que se muestra en pantalla,
loadedUsernames es un mapa que no se muestra en la pagina y que solo se usa para tener el recuento de qué usuarios ya se han cargado para no cargarlos de nuevo
*/

func InitialUserSearchPageModel(myUsername string, token []byte, searched string, client *http.Client) UserSearchPage {
	model := UserSearchPage{}
	model.client = client
	model.myUsername = myUsername
	model.token = token

	model.usernames = make([]string, 0)
	model.searchBar = textinput.New()
	model.searchBar.Placeholder = "Search user..."
	model.searchBar.Focus()
	model.onSearchBtn = true
	model.selectedUser = -1

	model.cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#FFF"))

	model.loadedUsernames = make(map[string]bool)
	model.canRequestMore = true
	model.pagesLoaded = 0
	model.searched = searched

	model.searchBar.SetValue(searched)
	return model
}

func (m UserSearchPage) Init() tea.Cmd {
	return nil
}

func (m UserSearchPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds = make([]tea.Cmd, 0, 4)
	var srchMsg tea.Cmd
	m.searchBar, srchMsg = m.searchBar.Update(msg)
	cmds = append(cmds, srchMsg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			return InitialHomeModel(m.myUsername, m.token, m.client), nil
		case "ctrl+c":
			return m, tea.Quit
		case "down":
			n := len(m.usernames)
			if m.selectedUser < n-1 {
				m.selectedUser++
			}

			if m.selectedUser != -1 {
				m.onSearchBtn = false
				m.searchBar.Blur()
			}

			if m.selectedUser == len(m.usernames)-1 && m.canRequestMore {
				cmds = append(cmds, GetUserMsg(m.pagesLoaded, m.searched, m.client))
			}

		case "up":
			if m.selectedUser >= 0 {
				m.selectedUser--
			}

			if m.selectedUser == -1 {
				m.onSearchBtn = true
				m.searchBar.Focus()
			}
		case "m":
			if m.token != nil && m.selectedUser >= 0 && m.usernames[m.selectedUser] != m.myUsername {
				return InitialChatPageModel(m.myUsername, m.token, m.client, m.usernames[m.selectedUser]),
					LoadChat(m.myUsername, m.token, m.usernames[m.selectedUser], m.client)
			}
		case "enter":
			if m.onSearchBtn {
				return InitialUserSearchPageModel(m.myUsername, m.token, m.searchBar.Value(), m.client), GetUserMsg(0, m.searchBar.Value(), m.client)
			}
		// else {
		// // Entrar pagina usu
		// }
		case "ctrl+r":
			cmd := GetUserMsg(0, "", m.client)
			return InitialUserSearchPageModel(m.myUsername, m.token, "", m.client), cmd
		}
	case message.ResetMsg:
		m.msg = ""
	case message.RequestLimitCooldown:
		m.canRequestMore = true
	case UsernamesMsg:
		users := msg
		if len(users) == 0 {
			m.canRequestMore = false
			cmds = append(cmds, message.SendTimedMessage(message.RequestLimitCooldown{}, 5*time.Second))
			m.msg = "No more users"
		} else {
			for _, u := range users {
				if _, ok := m.loadedUsernames[u]; !ok {
					m.usernames = append(m.usernames, u)
					m.loadedUsernames[u] = true
				}
			}

			// Si se ha recibido una lista incompleta con nombres (por ejemplo solo 2 cuando el tamaño de pagina es 5) -> hemos llegado al final de la lista
			// si se crean nuevos usuarios mientras un usuario esta en esta situacion, estos se crearan en esta misma pagina, por lo que no se pasa de pagina a menos que
			// el numero de usuarios sea multiplo de 5 o se haya recibido 5 en esta peticion (por si se recibe lista de usuarios ya cargados, pasar de pagina)
			if len(m.usernames)%usersPerReq == 0 || len(users) == usersPerReq {
				m.pagesLoaded++
			}

			if len(users) < usersPerReq {
				m.canRequestMore = false
				cmds = append(cmds, message.SendTimedMessage(message.RequestLimitCooldown{}, 5*time.Second))
				m.msg = "Reached end of user list"
			}
		}

	}

	if m.msg != "" {
		cmds = append(cmds, message.SendTimedMessage(message.ResetMsg{}, time.Second*5))
	}

	return m, tea.Batch(cmds...)
}

func (m UserSearchPage) View() string {
	var s string

	s = "User search\n\n"

	s += m.searchBar.View()

	start := m.selectedUser - int(math.Floor(float64(listSize)/2))
	end := start + listSize

	if start < 0 {
		end = listSize
	}

	n := len(m.usernames)

	if end > n {
		end = n
		start = end - listSize

	}
	if start < 0 {
		start = 0
	}

	s += "\n_________________________\n"
	for i := start; i < end; i++ {
		if m.selectedUser == i {
			s += m.cursorStyle.Render(m.usernames[i]) + "\n"
		} else {
			s += m.usernames[i] + "\n"
		}
	}

	for i := 0; i < listSize-end-start; i++ {
		s += "\n"
	}

	s += "‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾\n\n"
	s += "ctrl+r to refresh\n"
	s += "'m' to message user\n\n"

	if m.msg != "" {
		s += fmt.Sprintf("Info: %v\n\n", m.msg)
	}

	return s
}

func GetUserMsg(page int, username string, client *http.Client) func() tea.Msg {
	return func() tea.Msg {
		res, err := client.Get(fmt.Sprintf("https://127.0.0.1:10443/users?name=%v&page=%v&size=%v", username, page, usersPerReq))

		if err != nil {
			return fmt.Errorf("error en la petición")
		}

		users := make([]model.UserPublicData, usersPerReq)

		err = json.NewDecoder(res.Body).Decode(&users)

		if err != nil {
			return fmt.Errorf("error decodificando JSON")
		}

		r := UsernamesMsg{}

		for _, u := range users {
			r = append(r, u.Name)
		}

		return r
	}
}
