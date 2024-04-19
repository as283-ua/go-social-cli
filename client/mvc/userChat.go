package mvc

import (
	"bytes"
	"client/global"
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

	user   model.User
	client *http.Client
}

var saveKey = make([]byte, 32)

func InitialChatPageModel(user model.User, client *http.Client, username string) ChatPage {
	m := ChatPage{}
	m.client = client
	m.user = user

	m.username = username
	m.viewport = viewport.New(80, 12)
	m.chat = model.Chat{
		UserA:    user.Name,
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
			return InitialUserSearchPageModel(m.user, "", m.client), GetUserMsg(0, "", m.client)
		case "ctrl+c":
			m.SaveChat()
			return m, tea.Quit
		case "enter":
			if m.user.Token == nil {
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

				message := model.Message{Sender: m.user.Name, Message: strings.TrimSpace(m.textbox.Value()), Timestamp: time.Now()}

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
			return InitialChatPageModel(m.user, m.client, m.username),
				LoadChat(m.user.Name, m.user.Token, m.username, m.client)
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
			if message.Sender == m.user.Name {
				m.messagesStr += MessageToString(message, m.meStyle) + "\n"
			} else if message.Sender == m.username {
				m.messagesStr += MessageToString(message, m.otherStyle) + "\n"
			} else {
				panic(message)
			}
		}

		m.viewport.SetContent(m.messagesStr)
		m.viewport.GotoBottom()

		// m.msg = "Cargado chat"
	case error:
		m.msg = fmt.Sprintf("error. %v", msg)
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

	body := model.Message{Message: util.Encode64(util.Encrypt([]byte(m.textbox.Value()), m.chat.Key)), Sender: m.user.Name}

	bodyBytes := util.EncodeJSON(body)

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))

	if err != nil {
		return fmt.Errorf("error creando request")
	}

	req.Header.Add("Authorization", util.Encode64(m.user.Token))
	req.Header.Add("Username", m.user.Name)

	resp, err := m.client.Do(req)

	if err != nil {
		return fmt.Errorf("error conectando con el servidor")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la peticion %v", resp.StatusCode)
	}

	return nil
}

func writeSaveKey(keyPath string) error {
	pubKey := global.GetPublicKey()

	if pubKey == nil {
		return fmt.Errorf("error. no hay clave publica guardada")
	}

	rand.Read(saveKey)

	encKey, err := util.EncryptWithRSA(saveKey, pubKey)
	if err != nil {
		return err
	}

	file, err := os.Create(keyPath)
	if err != nil {
		return err
	}

	defer file.Close()
	file.Write(encKey)

	file2, err := os.Create(keyPath + ".pub")
	if err != nil {
		return err
	}

	defer file2.Close()
	file2.Write([]byte(fmt.Sprintf("%v", saveKey)))

	return nil
}

func getSaveKey(keyPath string) error {
	privKey := global.GetPrivateKey()
	var err error

	if privKey == nil {
		return fmt.Errorf("no hay saveKey")
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		return err
	}

	var encKey []byte = make([]byte, info.Size())

	file, err := os.Open(keyPath)

	if err != nil {
		return err
	}

	defer file.Close()
	file.Read(encKey)

	saveKey, err = util.DecryptWithRSA(encKey, privKey)
	if err != nil {
		return fmt.Errorf("clave aes %v\n clave privada %v\n clave encriptada %v", saveKey, privKey, encKey)
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
			privKey := global.GetPrivateKey()
			if privKey == nil {
				return fmt.Errorf("error obteniendo clave privada")
			}

			decoded, err := util.Decode64(unread[0].Message)
			if err != nil {
				return fmt.Errorf("error decodificando mensaje")
			}

			chat.Key, err = util.DecryptWithRSA(decoded, privKey)
			if err != nil {
				return fmt.Errorf("error decodificando mensaje")
			}

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
	err := getSaveKey(fmt.Sprintf("./chats/%s/%s.key", username, usernameOther))
	if err != nil {
		if os.IsNotExist(err) {
			// aun no hay clave para guardado
			return nil
		}
		return err
	}

	chatEnc, err := os.ReadFile(fmt.Sprintf("./chats/%s/%s.enc", username, usernameOther))

	if err != nil {
		if os.IsNotExist(err) {
			return message.FirstChatMsg{}
		}
		return err
	}

	chatJson, err := util.Decrypt(chatEnc, saveKey)

	if err != nil {
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
	chatJson, err := json.Marshal(m.chat)

	if err != nil {
		return fmt.Errorf("error json")
	}

	chatsPath := fmt.Sprintf("./chats/%s", m.user.Name)
	if _, err := os.Stat(chatsPath); os.IsNotExist(err) {
		err = os.Mkdir(chatsPath, fs.ModePerm)

		if err != nil {
			dir, _ := os.Getwd()
			return fmt.Errorf("error creando carpeta %s desde %s", chatsPath, dir)
		}
	}

	err = writeSaveKey(fmt.Sprintf("./chats/%s/%s.key", m.user.Name, m.username))
	if err != nil {
		return err
	}

	chatEnc := util.Encrypt(chatJson, saveKey)

	file, err := os.Create(fmt.Sprintf("%s/%s.enc", chatsPath, m.username))

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(chatEnc)

	if err != nil {
		return err
	}

	file, _ = os.Create(fmt.Sprintf("%s/%s.json", chatsPath, m.username))
	defer file.Close()
	file.Write(chatJson)

	return nil
}
