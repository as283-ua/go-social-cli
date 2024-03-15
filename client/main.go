package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"util"
	"util/models"
)

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	for {
		resp := registerCmdLine(client)
		if !resp.Ok {
			fmt.Println(resp.Msg)
			continue
		}

		resp = loginCmdLine(client)
		if !resp.Ok {
			fmt.Println(resp.Msg)
		}
	}
}

func registerCmdLine(client *http.Client) models.Resp {

	fmt.Print("Register\n\tUsuario: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	fmt.Print("\tPassword: ")
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	register := models.Credentials{User: strings.TrimRight(username, "\n"), Pass: strings.TrimRight(password, "\n")}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/register", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println(err)
	}

	r := util.DecodeJSON[models.Resp](resp.Body)
	fmt.Println(r)

	resp.Body.Close()
	return r
}

func loginCmdLine(client *http.Client) models.Resp {
	fmt.Print("Login\n\tUsuario: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	fmt.Print("\tPassword: ")
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	register := models.Credentials{User: strings.TrimRight(username, "\n"), Pass: strings.TrimRight(password, "\n")}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/login", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println(err)
	}

	r := util.DecodeJSON[models.Resp](resp.Body)
	fmt.Println(r)

	resp.Body.Close()
	return r
}
