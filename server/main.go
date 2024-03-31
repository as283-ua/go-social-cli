package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"server/handler"
	"server/logging"
	"server/middleware"
	"strconv"
	"strings"
	"syscall"
	"time"
	"util"
	"util/model"
)

var key []byte //clave para encriptar y desencriptar la base de datos, se introduce manualmente al arrancar el servidor

var data model.Database

// este metodo guarda la info de la base de datos en un archivo json sin encriptar para que podamos ver el contenido
func saveDatabaseJSON() {
	jsonData := util.EncodeJSON(data)

	err := os.WriteFile("db.json", jsonData, 0644)
	util.FailOnError(err)

	// fmt.Println("Base de datos guardada en", "db.json")
}

func saveDatabase() {
	jsonData := util.EncodeJSON(data)

	encryptedData := util.Encrypt(jsonData, key)

	err := os.WriteFile("db.enc", encryptedData, 0644)
	util.FailOnError(err)

	// fmt.Println("Base de datos guardada en", "db.enc")
}

func loadDatabase() error {
	encryptedData, err := os.ReadFile("db.enc")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("El archivo de la base de datos no existe.")
			data = model.Database{
				Users:            make(map[string]model.User),
				Groups:           make(map[string]model.Group),
				Posts:            make(map[int]model.Post),
				UserPosts:        make(map[string][]int),
				GroupPosts:       make(map[string][]int),
				GroupUsers:       make(map[string][]string),
				UserGroups:       make(map[string][]string),
				UserNames:        make([]string, 0),
				PendingCertLogin: make(map[string][]byte),
				PendingMessages:  make(map[string][]model.Message),
				NextPostId:       0,
			}
			return nil
		}

		return err
	}

	jsonData, err := util.Decrypt(encryptedData, key)

	if err != nil {
		return fmt.Errorf("clave incorrecta")
	}

	err = json.Unmarshal(jsonData, &data)

	if err != nil {
		return fmt.Errorf("clave incorrecta")
	}

	data.PendingCertLogin = make(map[string][]byte)
	fmt.Println("Base de datos cargada desde db.enc")
	return nil
}

func saveState(intervalo int) {
	ticker := time.NewTicker(time.Duration(intervalo) * time.Second)
	for {
		saveDatabase()
		saveDatabaseJSON()
		<-ticker.C
	}
}

func setupInterruptHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		fmt.Println("Guardando la base de datos")
		saveDatabase()
		saveDatabaseJSON()
		os.Exit(1)
	}()
}

func main() {
	fmt.Printf("Introduce la clave para desencriptar la base de datos: ")
	introducedKey, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)
	hash := sha256.Sum256([]byte(strings.TrimSpace(introducedKey)))
	key = hash[:]

	err = loadDatabase()
	if err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}
	setupInterruptHandler()

	intervalo := 30 //intervalo por defecto = 30 segundos
	if len(os.Args) == 2 {
		intervaloStr := os.Args[1]
		intervalo, err = strconv.Atoi(intervaloStr)
		util.FailOnError(err)
	}

	go saveState(intervalo) //multiplico por 1000 para que sean segundos

	router := http.NewServeMux()

	router.HandleFunc("POST /register", handler.RegisterHandler)
	router.HandleFunc("POST /login", handler.LoginHandler)
	router.HandleFunc("GET /login/cert", handler.GetLoginCertHandler)
	router.HandleFunc("POST /login/cert", handler.PostLoginCertHandler)
	router.HandleFunc("GET /users", handler.GetUserNamesHandler)
	router.Handle("POST /posts", middleware.Authorization(http.HandlerFunc(handler.CreatePostHandler)))
	router.HandleFunc("GET /posts", handler.GetPostsHandler)
	// router.Handle("GET /chat/{user}", middleware.Authorization(http.HandlerFunc(handler.ChatConnectionHandler)))
	router.Handle("POST /chat/{user}/message", middleware.Authorization(http.HandlerFunc(handler.SendMessageHandler)))
	router.Handle("GET /chat/{user}/message", middleware.Authorization(http.HandlerFunc(handler.GetPendingMessages)))
	router.Handle("GET /chat/{user}/pk", middleware.Authorization(http.HandlerFunc(handler.GetPubKeyHandler)))

	router.Handle("POST /noauth/chat/{user}/message", http.HandlerFunc(handler.SendMessageHandler))
	router.Handle("GET /noauth/chat/{user}/message", http.HandlerFunc(handler.GetPendingMessages))

	server := http.Server{
		Addr:    ":10443",
		Handler: middleware.InjectData(&data)(router),
	}

	fmt.Printf("Servidor escuchando en https://localhost:10443\n")
	util.FailOnError(server.ListenAndServeTLS("localhost.crt", "localhost.key"))
}
