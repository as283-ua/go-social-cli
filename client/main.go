package main

import (
	"client/mvc"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"util/model"

	tea "github.com/charmbracelet/bubbletea"
)

var UserName string

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	p := tea.NewProgram(mvc.InitialHomeModel(model.User{}, client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
