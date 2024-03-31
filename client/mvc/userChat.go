package mvc

import (
	"client/message"
	"fmt"
	"net/http"
	"strings"
	"util/model"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func MessageToString(m model.ChatMessage, senderStyle lipgloss.Style) string {
	return senderStyle.Render("@"+m.Sender) + "\n" + m.Message + "\n"
}

type ChatPage struct {
	username    string
	chat        model.Chat
	messagesStr string
	viewport    viewport.Model
	textbox     textarea.Model
	msg         string

	meStyle    lipgloss.Style
	otherStyle lipgloss.Style

	myUsername string
	client     *http.Client
	token      []byte
}

func InitialChatPageModel(myUsername string, token []byte, client *http.Client, username string) ChatPage {
	m := ChatPage{}
	m.client = client
	m.myUsername = myUsername
	m.token = token

	m.username = username
	m.viewport = viewport.New(80, 12)
	m.chat = model.Chat{
		UserA:    myUsername,
		UserB:    username,
		Messages: make([]model.ChatMessage, 0),
	}

	m.textbox = textarea.New()
	m.textbox.Focus()
	m.textbox.Placeholder = "Send a message..."
	m.textbox.Prompt = "┃ "
	m.textbox.CharLimit = 280
	m.textbox.ShowLineNumbers = false
	m.textbox.SetHeight(5)
	m.textbox.SetWidth(80)
	m.textbox.FocusedStyle.CursorLine = lipgloss.NewStyle()

	m.meStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8"))
	m.otherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#45f"))

	return m
}

func (m ChatPage) Init() tea.Cmd {
	return nil
}

func (m ChatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 2)

	m.viewport, cmds[0] = m.viewport.Update(msg)
	m.textbox, cmds[1] = m.textbox.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			return InitialUserSearchPageModel(m.myUsername, m.token, "", m.client), nil
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+s":
			if m.token == nil {
				m.msg = "No token. Can't post"
				break
			}

			if strings.TrimSpace(m.textbox.Value()) == "" {
				break
			}

			err := m.Send()
			if err != nil {
				m.msg = err.Error()
			} else {
				m.viewport.GotoBottom()

				message := model.ChatMessage{Sender: m.myUsername, Message: strings.TrimSpace(m.textbox.Value())}

				m.chat.Messages = append(m.chat.Messages, message)
				m.messagesStr += MessageToString(message, m.meStyle) + "\n"

				m.viewport.SetContent(m.messagesStr)
				m.textbox.Reset()
			}
		}
	case message.ReceiveMessageMsg:
		message := model.ChatMessage{Sender: msg.Sender, Message: strings.TrimSpace(msg.Message)}

		m.chat.Messages = append(m.chat.Messages, message)
		m.messagesStr += MessageToString(message, m.otherStyle) + "\n"

		m.viewport.SetContent(m.messagesStr)
	case message.ChatMsg:
		chat := msg
		m.chat = model.Chat(chat)
	}
	return m, tea.Batch(cmds...)
}

func (m ChatPage) View() string {
	var s string

	s = fmt.Sprintf("Chat with '%s'\n", m.username)

	s += "_________________________\n"
	s += m.viewport.View() + "\n"
	s += "‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾\n\n"
	s += fmt.Sprintf("Post as %s:\n", m.username)
	s += m.textbox.View() + "\n"
	s += "ctrl+s to post\n"

	if m.msg != "" {
		s += fmt.Sprintf("Info: %s\n\n", m.msg)
	}

	return s
}

func (m *ChatPage) Send() error {
	// m.client.
	return nil
}

func LoadChat(username string, usernameOther string) func() tea.Msg {
	return func() tea.Msg {
		return message.ChatMsg{
			UserA:    username,
			UserB:    usernameOther,
			Messages: make([]model.ChatMessage, 0),
		}
	}
}
