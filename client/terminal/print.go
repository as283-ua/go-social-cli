package terminal

import (
	"fmt"
	"strings"
	"util"

	"golang.org/x/term"
)

var options = []string{
	"1: Register",
	"2: Login",
	"3: Post",
	"4: All posts",
	"5: Create group",
	"6: Log out",
	"q: Quit",
}

type UiState struct {
	LoggedIn       bool
	SelectedOption int
}

var uiState = UiState{SelectedOption: -1}

func PrintUI(accion string) {
	for i, option := range options {
		if strings.Contains(option, accion) {
			uiState.SelectedOption = i
		}
	}

	width, height, err := term.GetSize(0)
	util.FailOnError(err)

	height -= 1

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if i == 0 || i == height-1 || j == 0 || j == width-1 {
				fmt.Print("\033[47m \033[49m")
				continue
			}

			if j == 2 && i-1 < len(options) {
				if i-1 == uiState.SelectedOption {
					fmt.Print("\033[30;47m")
				}
				fmt.Print(options[i-1])
				fmt.Print("\033[0m")
				j += len(options[i-1]) - 1
				continue
			}

			if j == width/5 {
				fmt.Print("\033[47m \033[49m")
				continue
			}

			fmt.Print(" ")
		}
	}
}
