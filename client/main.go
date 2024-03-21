package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"util"
	"util/models"
)

var token []byte
var UserName string

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	for {
		fmt.Print("Acciones disponibles:\n\t1: Register\n\t2: Login\n\t3: Post\n\t4: Todos posts\n\tq: Quit\nSeleccione la accion a realizar:")
		accion, err := bufio.NewReader(os.Stdin).ReadString('\n')
		util.FailOnError(err)

		accion = strings.TrimSpace(accion)

		switch accion {
		case "1":
			resp := registerCmdLine(client)
			if !resp.Ok {
				fmt.Println(resp.Msg)
				continue
			}
		case "2":
			resp := loginCmdLine(client)
			if !resp.Ok {
				fmt.Println(resp.Msg)
			}
		case "3":
			postPost(client)
		case "4":
		case "5":
			logOut()
		case "q":
			os.Exit(0)
		default:
			fmt.Print("Accion invalida, vuelva a intentarlo.\n\n")
			continue
		}
	}
}

func registerCmdLine(client *http.Client) models.Resp {

	fmt.Print("\nRegister\n\tUsuario: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	fmt.Print("\tPassword: ")
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	var publicKeyBytes []byte
	var privateKey *rsa.PrivateKey
	if _, err := os.Stat(fmt.Sprintf("%s.key", username)); err != nil {
		// no hay err -> el archivo no existe
		pk, err := rsa.GenerateKey(rand.Reader, 2048)
		privateKey = pk
		util.FailOnError(err)

		// writeECDSAKeyToFile(fmt.Sprintf("%s.key", username), privateKey)
		util.WriteRSAKeyToFile(fmt.Sprintf("%s.key", username), privateKey)
		publicKeyBytes = util.WritePublicKeyToFile(fmt.Sprintf("%s.pub", username), &privateKey.PublicKey)
	} else {
		privateKey = util.ReadRSAKeyFromFile(fmt.Sprintf("%s.key", username))
		publicKeyBytes = util.ReadPublicKeyBytesFromFile(fmt.Sprintf("%s.pub", username))
	}

	register := models.RegisterCredentials{User: strings.TrimSpace(username), Pass: strings.TrimRight(password, "\n"), PubKey: publicKeyBytes}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/register", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println(err)
	}

	r := util.DecodeJSON[models.Resp](resp.Body)
	if !r.Ok {
		fmt.Print("El usuario ya existe.\n\n")
	} else {
		util.DecryptWithRSA(util.Decode64(r.Msg), privateKey)
		token = r.Token
		UserName = strings.TrimSpace(username)

	}

	// fmt.Println(mensaje)

	// util.DecryptWithRSA(util.EncryptWithRSA([]byte("hola"), util.ParsePublicKey(publicKeyBytes)), privateKey)

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

	register := models.Credentials{User: strings.TrimSpace(username), Pass: strings.TrimRight(password, "\n")}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/login", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println(err)
	}

	r := util.DecodeJSON[models.Resp](resp.Body)
	defer resp.Body.Close()
	r.Msg = string(util.Decode64(r.Msg))
	fmt.Println(r)

	if !r.Ok {
		return r
	}

	token = r.Token
	UserName = strings.TrimSpace(username)

	return r
}

func postPost(client *http.Client) models.Resp {
	fmt.Print("Post\n\tContenido: ")
	http.NewRequest("POST", "https://localhost:10443/posts", nil)

	content, err := bufio.NewReader(os.Stdin).ReadString('\n')

	util.FailOnError(err)

	register := models.Post{Content: strings.TrimRight(content, "\n")}
	jsonBody := util.EncodeJSON(register)

	req, err := http.NewRequest("POST", "https://localhost:10443/posts", bytes.NewReader(jsonBody))
	util.FailOnError(err)
	req.Header.Add("content-type", "application/json")

	req.Header.Add("Authorization", util.Encode64(token))
	req.Header.Add("UserName", UserName)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return models.Resp{Ok: false, Msg: "Error en la peticion"}
	}

	r := util.DecodeJSON[models.Resp](resp.Body)
	r.Msg = string(util.Decode64(r.Msg))
	fmt.Println(r)

	resp.Body.Close()
	return r
}

func logOut() {
	token = nil
	UserName = ""
}
