package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"util"
	"util/model"
)

var token []byte
var UserName string

var options = []string{
	"1: Register",
	"2: Login",
	"3: Post",
	"4: All posts",
	"5: Create group",
	"6: Join group",
	"7: Block User (admins only)",
	"8: SSE Chat",
	"9: Log out",
	"q: Quit",
}

func printOptions() {
	fmt.Println("Opciones disponibles:")
	for _, option := range options {
		fmt.Println("\t" + option)
	}

}

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	for {
		printOptions()
		fmt.Print("Seleccione una accion: ")
		accion, err := bufio.NewReader(os.Stdin).ReadString('\n')
		util.FailOnError(err)

		accion = strings.TrimSpace(accion)

		switch accion {
		case "1":
			err := registerCmdLine(client)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "2":
			err := loginCmdLine(client)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "3":
			fmt.Println("\t1: General Post\n\t2: Group Post")
			option, err := bufio.NewReader(os.Stdin).ReadString('\n')
			util.FailOnError(err)
			switch option {
			case "1":
				postPost(client)
			case "2":
				//post in group
			}
		case "4":
			getPosts(client)
		case "5":
			fmt.Println("No implementado")
		case "6":
			fmt.Println("No implementado")
		case "7":
			fmt.Println("No implementado")
		case "8":
			fmt.Print("Usuario con el que desea chatear: ")
			user, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			testSSE(client, strings.TrimSpace(user))
		case "9":
			logOut()
		case "q":
			os.Exit(0)
		default:
			fmt.Print("Accion invalida, vuelva a intentarlo.\n\n")
			continue
		}
	}
}

func registerCmdLine(client *http.Client) error {

	fmt.Print("\nRegister\n\tUsuario: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)
	username = strings.TrimSpace(username)

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

	register := model.RegisterCredentials{User: strings.TrimSpace(username), Pass: strings.TrimSpace(password), PubKey: publicKeyBytes}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/register", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("error al hacer la peticion")
	}

	var r = model.Resp{}
	util.DecodeJSON(resp.Body, &r)
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
	return nil
}

func loginCmdLine(client *http.Client) error {
	fmt.Print("Login\n\tUsuario: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	fmt.Print("\tPassword: ")
	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)

	register := model.Credentials{User: strings.TrimSpace(username), Pass: strings.TrimRight(password, "\n")}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/login", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("error al hacer la peticion")
	}

	var r = model.Resp{}
	util.DecodeJSON(resp.Body, &r)
	defer resp.Body.Close()
	r.Msg = string(util.Decode64(r.Msg))
	fmt.Println(r)

	if !r.Ok {
		return fmt.Errorf("credenciales invalidas")
	}

	token = r.Token
	UserName = strings.TrimSpace(username)

	return nil
}

func postPost(client *http.Client) error {
	fmt.Print("Post\n\tContenido: ")
	http.NewRequest("POST", "https://localhost:10443/posts", nil)

	content, err := bufio.NewReader(os.Stdin).ReadString('\n')

	util.FailOnError(err)

	post := model.PostContent{Content: strings.TrimRight(content, "\n")}
	jsonBody := util.EncodeJSON(post)

	req, err := http.NewRequest("POST", "https://localhost:10443/posts", bytes.NewReader(jsonBody))
	util.FailOnError(err)
	req.Header.Add("content-type", "application/json")

	req.Header.Add("Authorization", util.Encode64(token))
	req.Header.Add("UserName", UserName)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	var r model.Resp
	util.DecodeJSON(resp.Body, &r)
	r.Msg = string(util.Decode64(r.Msg))
	fmt.Println(r)

	resp.Body.Close()
	return nil
}

func getPosts(client *http.Client) {
	req, err := http.NewRequest("GET", "https://localhost:10443/posts", nil)
	util.FailOnError(err)
	req.Header.Add("content-type", "application/json")

	// req.Header.Add("Authorization", util.Encode64(token))
	// req.Header.Add("UserName", UserName)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	var posts map[int]model.Post
	util.DecodeJSON(resp.Body, &posts)

	for _, v := range posts {
		fmt.Println(v)
	}

	resp.Body.Close()
}

func testSSE(client *http.Client, user string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:10443/chat/%s", user), nil)
	util.FailOnError(err)

	req.Header.Set("Accept", "text/event-stream")

	req.Header.Add("Authorization", util.Encode64(token))
	req.Header.Add("UserName", UserName)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			util.FailOnError(err)
		}

		if len(line) > 0 {
			fmt.Printf("Received: %s", line) // Remove "data: "
		}
		// serv debe enviar acabado en \n\n
		reader.ReadBytes('\n')
	}
}

func logOut() {
	token = nil
	UserName = ""
}
