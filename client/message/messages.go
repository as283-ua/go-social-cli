package message

import (
	"time"
	"util/model"

	tea "github.com/charmbracelet/bubbletea"
)

type RequestLimitCooldown struct{}
type ResetMsg struct{}
type ReceiveMessageMsg model.ChatMessage
type ChatMsg model.Chat

func SendTimedMessage(msg interface{}, t time.Duration) func() tea.Msg {
	return func() tea.Msg {
		timer := time.NewTimer(t)
		<-timer.C

		return msg
	}
}
