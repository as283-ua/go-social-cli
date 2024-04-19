package mvc

import (
	"bytes"
	"client/message"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"util"
	"util/model"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PostListModel struct {
	viewport viewport.Model
	posts    []string
	textbox  textarea.Model
	msg      string
	group    string

	client         *http.Client
	user           model.User
	pagesLoaded    int
	postsLoaded    map[int]bool
	canRequestMore bool
}

/*
Aclaracion sobre componentes del modelo:
	- PostsLoaded. Mapa de ids de posts que se han cargado para que al recibir una pagina, si hay posts que han hecho que se descuadre la paginacion, solo se añaden los nuevos

	- Can request more. Para evitar que se envian muchas peticiones aposta al llegar al final de la pagina, se fija un timer de 5 segundos que impide hacer peticiones de carga
*/

type PostsMsg []model.Post

const postsPerReq = 10

func comprobarAccesoGrupo(username, group string, token []byte, client *http.Client) bool {
	if group == "" {
		return true
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://127.0.0.1:10443/groups/%v/access", group), nil)
	req.Header.Add("Username", username)
	req.Header.Add("Authorization", util.Encode64(token))

	if err != nil {
		return false
	}

	resp, err := client.Do(req)

	if err != nil {
		return false
	}

	var objResp model.Resp
	util.DecodeJSON(resp.Body, &objResp)

	return objResp.Ok
}

func InitialPostListModel(user model.User, group string, client *http.Client) (PostListModel, error) {
	m := PostListModel{}

	m.group = group
	if !comprobarAccesoGrupo(user.Name, group, user.Token, client) {
		return m, fmt.Errorf("acceso denegado")
	}

	m.client = client
	m.user = user
	m.canRequestMore = true
	m.postsLoaded = make(map[int]bool)

	m.viewport = viewport.New(80, 12)

	m.posts = make([]string, 5)

	m.textbox = textarea.New()
	m.textbox.Focus()
	m.textbox.Placeholder = "Send a message..."
	m.textbox.Prompt = "┃ "
	m.textbox.CharLimit = 280
	m.textbox.ShowLineNumbers = false
	m.textbox.SetHeight(5)
	m.textbox.SetWidth(80)

	// Remove cursor line styling
	m.textbox.FocusedStyle.CursorLine = lipgloss.NewStyle()

	return m, nil
}

func GetPostsMsg(page int, group, username string, token []byte, client *http.Client) func() tea.Msg {
	return func() tea.Msg {
		var url string
		if group == "" {
			url = fmt.Sprintf("https://127.0.0.1:10443/posts?page=%v&size=%v", page, postsPerReq)
		} else {
			url += fmt.Sprintf("https://127.0.0.1:10443/groups/%v/posts?page=%v&size=%v", group, page, postsPerReq)
		}

		req, _ := http.NewRequest("GET", url, nil)
		if group != "" {
			req.Header.Add("Username", username)
			req.Header.Add("Authorization", util.Encode64(token))
		}

		res, err := client.Do(req)

		if err != nil {
			return nil
		}

		if res.StatusCode != 200 {
			return fmt.Errorf(res.Status)
		}

		posts := PostsMsg{
			model.Post{},
			model.Post{},
			model.Post{},
			model.Post{},
			model.Post{},
		}

		json.NewDecoder(res.Body).Decode(&posts)

		return posts
	}
}

func (m PostListModel) Init() tea.Cmd {
	// llamar al servidor para conseguir 5 posts
	return nil
}

func (m PostListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		postTboxCmd tea.Cmd
		viewPortCmd tea.Cmd
	)

	if m.user.Name != "" {
		m.textbox, postTboxCmd = m.textbox.Update(msg)
	}
	m.viewport, viewPortCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			return InitialHomeModel(m.user, m.client), nil
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+r":
			m, _ := InitialPostListModel(m.user, m.group, m.client)
			return m, GetPostsMsg(0, m.group, m.user.Name, m.user.Token, m.client)
		case "ctrl+s":
			if m.user.Token == nil {
				m.msg = "No token. Can't post"
				break
			}

			postId, err := m.PublishPost()
			if err != nil {
				m.msg = err.Error()
			} else {
				m.msg = "Posted!"

				m.viewport.GotoTop()

				m.postsLoaded[postId] = true

				newPost := make([]string, 1)
				newPost[0] = InitialPost(model.Post{Content: m.textbox.Value(), Author: m.user.Name}).View()
				m.posts = slices.Concat(newPost, m.posts)

				m.viewport.SetContent(strings.Join(m.posts, ""))
				m.textbox.Reset()
			}
		case "down":
			if m.viewport.AtBottom() && m.canRequestMore {
				return m, GetPostsMsg(m.pagesLoaded, m.group, m.user.Name, m.user.Token, m.client)
			}
		}
	case message.ResetMsg:
		m.msg = ""
	case message.RequestLimitCooldown:
		m.canRequestMore = true
	case PostsMsg:

		if len(msg) == 0 {
			m.msg = "No new posts"
			m.canRequestMore = false
			return m, tea.Batch(postTboxCmd, viewPortCmd, message.SendTimedMessage(message.RequestLimitCooldown{}, 5*time.Second), message.SendTimedMessage(message.ResetMsg{}, 5*time.Second))
		}

		postRender := InitialPost(model.Post{})
		for _, post := range msg {
			if _, ok := m.postsLoaded[post.Id]; ok {
				continue
			}

			m.postsLoaded[post.Id] = true

			postRender.post = post
			m.posts = append(m.posts, postRender.View())
		}

		m.msg = "Loaded posts"
		m.pagesLoaded++
		m.viewport.SetContent(strings.Join(m.posts, ""))
	}

	if m.msg != "" {
		return m, tea.Batch(postTboxCmd, viewPortCmd, message.SendTimedMessage(message.ResetMsg{}, 5*time.Second))
	}

	return m, tea.Batch(postTboxCmd, viewPortCmd)
}

func (m PostListModel) View() string {
	var s string

	if m.group == "" {
		s = "Posts\n\n"
	} else {

		s = m.group + " posts\n\n"
	}

	s += "_________________________\n"
	s += m.viewport.View() + "\n"

	s += "‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾\n\n"

	if m.user.Token != nil {
		s += fmt.Sprintf("Post as %s:\n", m.user.Name)
		s += m.textbox.View() + "\n"
		s += "ctrl+s to post\n"
	}

	s += "ctrl+r to refresh\n\n"

	if m.msg != "" {
		s += fmt.Sprintf("Info: %s\n\n", m.msg)
	}

	return s
}

func (m PostListModel) PublishPost() (int, error) {
	post := model.Post{Content: m.textbox.Value()}
	postBytes := util.EncodeJSON(post)
	var url string

	if m.group == "" {
		url = "https://127.0.0.1:10443/posts"
	} else {
		url = fmt.Sprintf("https://127.0.0.1:10443/groups/%v/post", m.group)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(postBytes))
	req.Header.Add("Username", m.user.Name)
	req.Header.Add("Authorization", util.Encode64(m.user.Token))
	res, err := m.client.Do(req)

	if err != nil {
		return -1, fmt.Errorf("error conectando con el servidor")
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return -1, fmt.Errorf("Login incorrecto")
	}

	resp := model.Resp{}
	err = util.DecodeJSON(res.Body, &resp)

	if err != nil {
		return -1, fmt.Errorf(fmt.Sprintf("error decodificando JSON: %v", resp))
	}

	if !resp.Ok {
		return -1, fmt.Errorf(resp.Msg)
	}

	return strconv.Atoi(resp.Msg)
}
