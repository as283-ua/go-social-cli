package mvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
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

	client         *http.Client
	username       string
	token          []byte
	pagesLoaded    int
	canRequestMore bool
}

type TimerResetMsg struct{}
type TimerCooldown struct{}

type PostsMsg []model.Post

func InitialPostListModel(username string, token []byte, client *http.Client) PostListModel {
	m := PostListModel{}

	m.client = client
	m.username = username
	m.token = token
	m.canRequestMore = true

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

	return m
}

func GetPostsMsg(page int, client *http.Client) func() tea.Msg {
	return func() tea.Msg {
		res, err := client.Get(fmt.Sprintf("https://127.0.0.1:10443/posts?page=%v&size=5", page))

		if err != nil {
			return nil
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

func resetAfterTime(t time.Duration) func() tea.Msg {
	return func() tea.Msg {
		timer := time.NewTimer(t)
		<-timer.C

		return TimerResetMsg{}
	}
}

func startCoolDown(t time.Duration) func() tea.Msg {
	return func() tea.Msg {
		timer := time.NewTimer(t)
		<-timer.C

		return TimerCooldown{}
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

	if m.username != "" {
		m.textbox, postTboxCmd = m.textbox.Update(msg)
	}
	m.viewport, viewPortCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			return InitialHomeModel(m.username, m.token, m.client), nil
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+s":
			if m.token == nil {
				m.msg = "No token. Can't post"
				break
			}

			err := m.PublishPost()
			if err != nil {
				m.msg = err.Error()
			} else {
				m.msg = "Posted!"

				newPost := make([]string, 1)
				newPost[0] = InitialPost(model.Post{Content: m.textbox.Value(), Author: m.username}).View()
				m.posts = slices.Concat(newPost, m.posts)

				m.viewport.SetContent(strings.Join(m.posts, ""))
				m.textbox.Reset()
			}
		case "down":
			if m.viewport.YOffset+5 > m.viewport.Height && m.canRequestMore {
				return m, GetPostsMsg(m.pagesLoaded, m.client)
			}
		}
	case TimerResetMsg:
		m.msg = ""
	case TimerCooldown:
		m.canRequestMore = true
	case PostsMsg:

		if len(msg) == 0 {
			m.msg = "No new posts"
			m.canRequestMore = false
			return m, tea.Batch(postTboxCmd, viewPortCmd, resetAfterTime(5*time.Second), startCoolDown(5*time.Second))
		}

		postRender := InitialPost(model.Post{})
		for _, post := range msg {
			postRender.post = post
			m.posts = append(m.posts, postRender.View())
		}

		m.msg = "Loaded posts"
		m.pagesLoaded++
		m.viewport.SetContent(strings.Join(m.posts, ""))
	}

	if m.msg != "" {
		return m, tea.Batch(postTboxCmd, viewPortCmd, resetAfterTime(5*time.Second))
	}

	return m, tea.Batch(postTboxCmd, viewPortCmd)
}

func (m PostListModel) View() string {
	s := "Posts\n\n"

	s += "_________________________\n"
	s += m.viewport.View() + "\n"

	s += "‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾\n\n"

	if m.username != "" {
		s += fmt.Sprintf("Post as %s:\n", m.username)
		s += m.textbox.View() + "\n"
	}

	if m.msg != "" {
		s += fmt.Sprintf("Info: %s\n\n", m.msg)
	}

	s += "ctrl+s to post\n\n"

	return s
}

func (m PostListModel) PublishPost() error {
	post := model.Post{Content: m.textbox.Value()}
	postBytes := util.EncodeJSON(post)
	req, _ := http.NewRequest("POST", "https://127.0.0.1:10443/posts", bytes.NewReader(postBytes))
	req.Header.Add("Username", m.username)
	req.Header.Add("Authorization", util.Encode64(m.token))
	res, err := m.client.Do(req)

	if err != nil {
		return fmt.Errorf("error conectando con el servidor")
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("Login incorrecto")
	}

	resp := model.Resp{}
	err = util.DecodeJSON(res.Body, &resp)

	if err != nil {
		return fmt.Errorf("error decodificando JSON")
	}

	if !resp.Ok {
		return fmt.Errorf(resp.Msg)
	}

	return nil
}
