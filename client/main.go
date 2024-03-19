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

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	for {
		fmt.Print("Acciones disponibles:\n\t1: Register\n\t2: Login\n\tq: Quit\nSeleccione la accion a realizar:")
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

	register := models.RegisterCredentials{User: strings.TrimRight(username, "\n"), Pass: strings.TrimRight(password, "\n"), PubKey: publicKeyBytes}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/register", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println(err)
	}

	r := util.DecodeJSON[models.Resp](resp.Body)
	if r.Ok == false {
		fmt.Print("El usuario ya existe.\n\n")
	} else {
		util.DecryptWithRSA(util.Decode64(r.Msg), privateKey)
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

	register := models.Credentials{User: strings.TrimRight(username, "\n"), Pass: strings.TrimRight(password, "\n")}
	jsonBody := util.EncodeJSON(register)

	resp, err := client.Post("https://localhost:10443/login", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println(err)
	}

	r := util.DecodeJSON[models.Resp](resp.Body)
	r.Msg = string(util.Decode64(r.Msg))
	fmt.Println(r)

	resp.Body.Close()
	return r
}
