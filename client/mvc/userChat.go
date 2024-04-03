package mvc

import (
	"bytes"
	"client/message"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
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
	return fmt.Sprintf("%s - %s\n%s\n", senderStyle.Render("@"+m.Sender), m.Timestamp.Format("2 Jan 2006 15:04:05"), m.Message)
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

				message := model.Message{Sender: m.myUsername, Message: strings.TrimSpace(m.textbox.Value()), Timestamp: time.Now()}

				m.chat.Messages = append(m.chat.Messages, message)
				m.messagesStr += MessageToString(message, m.meStyle) + "\n"

				m.viewport.SetContent(m.messagesStr)
				m.viewport.GotoBottom()

				m.textbox.Reset()
			}
		case "ctrl+s":
			err := m.SaveChat()

			if err != nil {
				m.msg = err.Error()
			} else {
				m.msg = "Chat guardado"
			}
		case "ctrl+r":
			return InitialChatPageModel(m.myUsername, m.token, m.client, m.username),
				LoadChat(m.myUsername, m.token, m.username, m.client)
		}
	case message.ReceiveMessageMsg:
		message := model.Message(msg)

		m.chat.Messages = append(m.chat.Messages, message)
		m.messagesStr += MessageToString(message, m.otherStyle) + "\n"

		m.viewport.SetContent(m.messagesStr)

		m.msg = "Recibido mensaje"
	case message.ChatMsg:
		m.chat = model.Chat(msg)

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
		m.viewport.GotoBottom()

		m.msg = "Cargado chat"
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

	body := model.Message{Message: util.Encode64(util.Encrypt([]byte(m.textbox.Value()), m.chat.Key)), Sender: m.myUsername}

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

func SendKey(username string, token []byte, usernameOther string, client *http.Client) ([]byte, error) {
	resp, err := client.Get(fmt.Sprintf("https://localhost:10443/chat/%s/pubkey", usernameOther))

	if err != nil {
		return nil, fmt.Errorf("error conectando con servidor para conseguir clave publica de %s", usernameOther)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error leyendo body")
	}

	pubkeybytes, err := util.Decode64(string(body))

	if err != nil {
		return nil, fmt.Errorf("error decodificando de base 64")
	}

	var aeskey = make([]byte, 32)
	rand.Read(aeskey)

	pubkey := util.ParsePublicKey(pubkeybytes)

	encryptedKey, err := util.EncryptWithRSA(aeskey, pubkey)

	if err != nil {
		return nil, fmt.Errorf("error encriptando la clave aes")
	}

	url := fmt.Sprintf("https://localhost:10443/chat/%s/message", usernameOther)

	bodyReq := make(map[string]string)
	bodyReq["Message"] = util.Encode64(encryptedKey)
	_, err = util.Decode64(bodyReq["Message"])

	if err != nil {
		return nil, fmt.Errorf("decodificando lo que acabas de codificar")
	}
	bodyBytes := util.EncodeJSON(bodyReq)

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))

	if err != nil {
		return nil, fmt.Errorf("error creando request")
	}

	req.Header.Add("Authorization", util.Encode64(token))
	req.Header.Add("Username", username)

	resp, err = client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error conectando con servidor para enviar clave simetrica a %s", usernameOther)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %v", resp.StatusCode)
	}

	return aeskey, nil
}

func LoadChat(username string, token []byte, usernameOther string, client *http.Client) func() tea.Msg {
	return func() tea.Msg {
		chat := model.Chat{UserA: username, UserB: usernameOther, Messages: make([]model.Message, 0)}
		unread := make([]model.Message, 0)
		loadMsg := LoadSavedChat(username, usernameOther)
		downloadMsg := DownloadUnread(username, token, usernameOther, client)

		// opciones: ambos vacios -> primera vez que nos contactamos username y yo, hay que enviar clave aes con rsa
		// hay nuevos mensajes pero no chat guardado, primer mensaje debe contener clave aes porque el otro ha iniciado chat
		// 		habria que ver que hacer si se borra chat manualmente, porque el nuevo mensaje no sera una clave aes
		// chat cargado pero sin mensajes nuevos -> todo gucci
		// chat cargado Y nuevos mensajes -> unir mensajes y devolver como message.ChatMsg

		var (
			prevChat    = false
			newMessages = false
		)

		switch loadMsg := loadMsg.(type) {
		case error:
			return loadMsg
		case message.FirstChatMsg:
			prevChat = false
		case message.ChatMsg:
			chat = model.Chat(loadMsg)
			prevChat = true
		}

		switch downloadMsg := downloadMsg.(type) {
		case error:
			return loadMsg
		case message.UnreadMsg:
			unread = downloadMsg
		}

		newMessages = len(unread) != 0

		if !prevChat && !newMessages {
			aeskey, err := SendKey(username, token, usernameOther, client)
			if err != nil {
				return err
			}
			chat.Key = aeskey
		} else if prevChat && newMessages {
			for i, message := range unread {
				decoded, err := util.Decode64(message.Message)

				if err != nil {
					return fmt.Errorf("error decodificando")
				}

				msgBytes, err := util.Decrypt(decoded, chat.Key)
				if err != nil {
					return fmt.Errorf("error descifrando con la clave simetrica")
				}
				unread[i].Message = string(msgBytes)
			}
			chat.Messages = slices.Concat(chat.Messages, unread)
		} else if newMessages { // chat nuevo iniciado por otro usuario
			privKey, err := util.ReadRSAKeyFromFile(fmt.Sprintf("%s.key", username))
			if err != nil {
				return fmt.Errorf("error obteniendo clave privada")
			}

			decoded, err := util.Decode64(unread[0].Message)
			if err != nil {
				return fmt.Errorf("error decodificando mensaje")
			}

			chat.Key = util.DecryptWithRSA(decoded, privKey)

			unread = unread[1:]

			for i, message := range unread {
				decoded, err := util.Decode64(message.Message)

				if err != nil {
					return fmt.Errorf("error decodificando")
				}

				msgBytes, err := util.Decrypt(decoded, chat.Key)
				if err != nil {
					return fmt.Errorf("error descifrando con la clave simetrica")
				}
				unread[i].Message = string(msgBytes)
			}

			chat.Messages = unread
		}

		return message.ChatMsg(chat)
	}
}

func LoadSavedChat(username string, usernameOther string) tea.Msg {
	chat := model.Chat{}
	chatJson, err := os.ReadFile(fmt.Sprintf("./chats/%s/%s.json", username, usernameOther))

	if err != nil {
		if os.IsNotExist(err) {
			return message.FirstChatMsg{}
		}
		return err
	}

	err = json.Unmarshal(chatJson, &chat)

	if err != nil {
		return err
	}

	return message.ChatMsg(chat)
}

func DownloadUnread(username string, token []byte, usernameOther string, client *http.Client) tea.Msg {
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
	return message.UnreadMsg(body)
}

func (m *ChatPage) SaveChat() error {
	if m.chat.Key == nil {
		return nil
	}
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
