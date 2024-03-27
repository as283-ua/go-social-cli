package mvc

import (
	"fmt"
	"strings"
	"util/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PostModel struct {
	post       model.Post
	userStyle  lipgloss.Style
	groupStyle lipgloss.Style
}

func InitialPost(post model.Post) PostModel {
	return PostModel{
		post:       post,
		userStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8")),
		groupStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#45f")),
	}
}

func (m PostModel) Init() tea.Cmd {
	return nil
}

func (m PostModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (m PostModel) View() string {
	s := "@" + m.userStyle.Render(m.post.Author)
	if m.post.Group != "" {
		s += fmt.Sprintf(" [%s]", m.groupStyle.Render(m.post.Group))
	}
	s += "\n"

	contentWords := strings.Split(m.post.Content, " ")

	curLen := 0
	maxLen := 60
	for _, word := range contentWords {
		wordLen := len(word)
		if wordLen >= maxLen {
			s += "\n" + word + "\n"
			curLen = 0
			continue
		}

		if curLen+wordLen > maxLen {
			s += "\n" + word
			curLen = wordLen
			continue
		}

		s += " " + word
		curLen += wordLen
	}

	s += "\n\n"

	return s
}
