package mvc

import (
	"bytes"
	"client/message"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
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

func MessageToString(m model.Message, senderStyle lipgloss.Style) string {
	return senderStyle.Render("@"+m.Sender) + " - " + m.Timestamp.Format("2 Jan 2006 15:04:05") + "\n" + m.Message + "\n"
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
		Messages: make([]model.Message, 0),
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
			m.SaveChat()
			return InitialUserSearchPageModel(m.myUsername, m.token, "", m.client), GetUserMsg(0, "", m.client)
		case "ctrl+c":
			m.SaveChat()
			return m, tea.Quit
		case "enter":
			if m.token == nil {
				m.msg = "Sin token. No se pudo enviar el mensaje"
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

				message := model.Message{Sender: m.myUsername, Message: strings.TrimSpace(m.textbox.Value()), Timestamp: time.Now()}

				m.chat.Messages = append(m.chat.Messages, message)
				m.messagesStr += MessageToString(message, m.meStyle) + "\n"

				m.viewport.SetContent(m.messagesStr)
				m.textbox.Reset()

				m.msg = "Enviado mensaje"
			}
		case "ctrl+s":
			err := m.SaveChat()

			if err != nil {
				m.msg = err.Error()
			} else {
				m.msg = "Chat guardado"
			}
		}
	case message.ReceiveMessageMsg:
		message := model.Message(msg)

		m.chat.Messages = append(m.chat.Messages, message)
		m.messagesStr += MessageToString(message, m.otherStyle) + "\n"

		m.viewport.SetContent(m.messagesStr)
		m.msg = "Recibido mensaje"
	case message.ChatMsg:
		if len(m.chat.Messages) == 0 {
			m.chat.Messages = msg.Messages
		} else {
			// primero el chat recien cargado y luego los mensajes no leidos descargados de antes (deberian ser mas nuevos)
			m.chat.Messages = slices.Concat(m.chat.Messages, msg.Messages)
			m.messagesStr = ""
		}

		for _, message := range m.chat.Messages {
			if message.Sender == m.myUsername {
				m.messagesStr += MessageToString(message, m.meStyle) + "\n"
			} else if message.Sender == m.username {
				m.messagesStr += MessageToString(message, m.otherStyle) + "\n"
			} else {
				panic(message)
			}
		}

		m.viewport.SetContent(m.messagesStr)
		m.msg = "Cargado chat de archivo local"
	case message.UnreadMsg:
		m.chat.Messages = slices.Concat(m.chat.Messages, msg)

		for _, message := range msg {
			m.messagesStr += MessageToString(message, m.meStyle) + "\n"
		}

		m.viewport.SetContent(m.messagesStr)

		m.msg = fmt.Sprintf("Descargado %v nuevos mensajes", len(msg))
	case error:
		m.msg = fmt.Sprintf("error. %s", msg.Error())
	}
	return m, tea.Batch(cmds...)
}

func (m ChatPage) View() string {
	var s string

	s = fmt.Sprintf("Chat with '%s'\n", m.username)

	s += "_________________________\n"
	s += m.viewport.View() + "\n"
	s += "‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾\n\n"
	s += m.textbox.View() + "\n"
	s += "ctrl+s to post\n"

	if m.msg != "" {
		s += fmt.Sprintf("Info: %s\n\n", m.msg)
	}

	return s
}

func (m *ChatPage) Send() error {
	url := fmt.Sprintf("https://localhost:10443/chat/%s/message", m.username)

	body := make(map[string]string)
	body["Message"] = m.textbox.Value()

	bodyBytes := util.EncodeJSON(body)

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))

	if err != nil {
		return fmt.Errorf("error creando request")
	}

	req.Header.Add("Authorization", util.Encode64(m.token))
	req.Header.Add("Username", m.myUsername)

	resp, err := m.client.Do(req)

	if err != nil {
		return fmt.Errorf("error conectando con el servidor")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la peticion %v", resp.StatusCode)
	}

	return nil
}

func LoadChat(username string, usernameOther string) func() tea.Msg {
	return func() tea.Msg {
		chat := model.Chat{}
		chatJson, err := os.ReadFile(fmt.Sprintf("./chats/%s/%s.json", username, usernameOther))

		if err != nil {
			return err
		}

		err = json.Unmarshal(chatJson, &chat)

		if err != nil {
			return err
		}

		return chat
	}
}

func (m *ChatPage) SaveChat() error {
	chatJson, err := json.Marshal(m.chat)

	if err != nil {
		return fmt.Errorf("error json")
	}

	chatsPath := fmt.Sprintf("./chats/%s", m.myUsername)
	if _, err := os.Stat(chatsPath); os.IsNotExist(err) {
		err = os.Mkdir(chatsPath, fs.ModePerm)

		if err != nil {
			dir, _ := os.Getwd()
			return fmt.Errorf("error creando carpeta %s desde %s", chatsPath, dir)
		}
	}

	// encriptar chat aqui
	file, err := os.Create(fmt.Sprintf("%s/%s.json", chatsPath, m.username))

	if err != nil {
		// return fmt.Errorf("error creando archivo %s", fmt.Sprintf("%s/%s.json", chatsPath, m.username))
		return err
	}

	defer file.Close()

	_, err = file.Write(chatJson)

	if err != nil {
		return err
	}

	return nil
}

func DownloadUnread(username string, token []byte, usernameOther string, client *http.Client) func() tea.Msg {
	return func() tea.Msg {
		url := fmt.Sprintf("https://localhost:10443/chat/%s/message", usernameOther)

		req, err := http.NewRequest("GET", url, &bytes.Reader{})

		if err != nil {
			return err
		}

		req.Header.Add("Authorization", util.Encode64(token))
		req.Header.Add("Username", username)

		resp, err := client.Do(req)

		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status: %v", resp.StatusCode)
		}

		var body = make([]model.Message, 0)
		util.DecodeJSON(resp.Body, &body)
		return body
	}
}
