package message

import (
	"time"
	"util/model"

	tea "github.com/charmbracelet/bubbletea"
)

type RequestLimitCooldown struct{}
type ResetMsg struct{}
type ReceiveMessageMsg model.Message
type ChatMsg model.Chat
type UnreadMsg []model.Message
type FirstChatMsg struct{}

func SendTimedMessage(msg interface{}, t time.Duration) func() tea.Msg {
	return func() tea.Msg {
		timer := time.NewTimer(t)
		<-timer.C

		return msg
	}
}
