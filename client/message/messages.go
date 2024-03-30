package message

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RequestLimitCooldown struct{}
type ResetMsg struct{}

func SendTimedMessage(msg interface{}, t time.Duration) func() tea.Msg {
	return func() tea.Msg {
		timer := time.NewTimer(t)
		<-timer.C

		return msg
	}
}
