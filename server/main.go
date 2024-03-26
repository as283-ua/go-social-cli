package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"server/repository"
	"strconv"
	"strings"
	"syscall"
	"time"
	"util"
	"util/model"

	"golang.org/x/crypto/argon2"
)

func NewTupleAlphabeticOrder(a, b string) model.UserChat {
	if a <= b {
		return model.UserChat{First: a, Second: b}
	}
	return model.UserChat{First: b, Second: a}
}

var key []byte //clave para encriptar y desencriptar la base de datos, se introduce manualmente al arrancar el servidor

// BD Principal
type Database struct {
	Users      map[string]model.User
	Groups     map[string]model.Group
	Posts      map[int]model.Post
	UserPosts  map[string][]int
	GroupPosts map[string][]int
	GroupUsers map[string][]string
	UserGroups map[string][]string
	UserNames  []string
}

var pendingCertLogin = make(map[string][]byte)

var data Database

// este metodo guarda la info de la base de datos en un archivo json sin encriptar para que podamos ver el contenido
func saveDatabaseJSON() {
	jsonData := util.EncodeJSON(data)

	err := os.WriteFile("db.json", jsonData, 0644)
	util.FailOnError(err)

	fmt.Println("Base de datos guardada en", "db.json")
}

func saveDatabase() {
	jsonData := util.EncodeJSON(data)

	encryptedData := util.Encrypt(jsonData, key)

	err := os.WriteFile("db.enc", encryptedData, 0644)
	util.FailOnError(err)

	fmt.Println("Base de datos guardada en", "db.enc")
}

func loadDatabase() error {
	encryptedData, err := os.ReadFile("db.enc")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("El archivo de la base de datos no existe.")
			data = Database{
				Users:      make(map[string]model.User),
				Groups:     make(map[string]model.Group),
				Posts:      make(map[int]model.Post),
				UserPosts:  make(map[string][]int),
				GroupPosts: make(map[string][]int),
				GroupUsers: make(map[string][]string),
				UserGroups: make(map[string][]string),
				UserNames:  make([]string, 0),
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

var logger = util.GetLogger()

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token, err := util.Decode64(req.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			util.FailOnError(err)
			return
		}

		if !validarToken(req.Header.Get("UserName"), string(token)) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func main() {

	fmt.Printf("Introduce la clave para desencriptar la base de datos: ")
	introducedKey, err := bufio.NewReader(os.Stdin).ReadString('\n')
	util.FailOnError(err)
	hash := sha256.Sum256([]byte(strings.TrimSpace(introducedKey)))
	key = hash[:]

	err = loadDatabase()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	setupInterruptHandler()

	intervalo := 30 //intervalo por defecto = 30 segundos
	if len(os.Args) == 2 {
		intervaloStr := os.Args[1]
		intervalo, err = strconv.Atoi(intervaloStr)
		util.FailOnError(err)
	}

	fmt.Print(intervalo)

	go saveState(intervalo) //multiplico por 1000 para que sean segundos

	server := http.Server{
		Addr: ":10443",
	}

	router := http.NewServeMux()

	router.HandleFunc("POST /register", registerHandler)
	router.HandleFunc("POST /login", loginHandler)
	router.HandleFunc("GET /login/cert", getLoginCertHandler)
	router.HandleFunc("POST /login/cert", postLoginCertHandler)
	router.HandleFunc("GET /users", usersHandler)
	router.Handle("POST /posts", Authorization(http.HandlerFunc(postsHandler)))
	router.HandleFunc("GET /posts", getPostsHandler)
	router.Handle("GET /chat/{user}", Authorization(http.HandlerFunc(chatHandler)))
	router.Handle("GET /noauth/chat/{user}", http.HandlerFunc(chatHandler))
	router.Handle("POST /chat/{user}/message", Authorization(http.HandlerFunc(sendMessageHandler)))
	router.Handle("GET /chat/{user}/pk", Authorization(http.HandlerFunc(getPkHandler)))

	server.Handler = router
	fmt.Printf("Servidor escuchando en https://localhost:10443\n")
	util.FailOnError(server.ListenAndServeTLS("localhost.crt", "localhost.key"))
}

func usersHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := req.URL.Query()
	pageStr := query.Get("page")
	sizeStr := query.Get("size")
	page := 0
	size := len(data.UserNames)

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		util.FailOnError(err)
		page = p
	}
	if sizeStr != "" {
		s, err := strconv.Atoi(sizeStr)
		util.FailOnError(err)
		size = s
	}

	start := page * size
	end := (page + 1) * size

	if end >= len(data.UserNames) {
		end = len(data.UserNames)
	}

	err := json.NewEncoder(w).Encode(data.UserNames[start:end])
	util.FailOnError(err)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	logger.Info("Register handler")
	w.Header().Set("Content-Type", "application/json")

	var register model.RegisterCredentials

	util.DecodeJSON(req.Body, &register)
	if register.User == "" || register.Pass == "" || register.PubKey == nil {
		response(w, false, "Campos vacíos", nil)
		return
	}

	// logger.Info(fmt.Sprintf("Registro: %v\n", register))

	w.Header().Set("Content-Type", "application/json")

	_, ok := data.Users[register.User]
	if ok {
		response(w, false, "Usuario ya registrado", nil)
		return
	}

	u := model.User{}
	u.Name = register.User
	u.Salt = make([]byte, 16)
	rand.Read(u.Salt)
	password := register.Pass

	u.Hash = argon2.Key([]byte(password), u.Salt, 3, 32*1024, 4, 32)

	u.Seen = time.Now()
	u.Token = make([]byte, 16)
	rand.Read(u.Token)

	u.PubKey = register.PubKey
	data.Users[u.Name] = u
	data.UserNames = append(data.UserNames, u.Name)

	encryptedMsg := util.EncryptWithRSA([]byte("Bienvenido a la red social"), util.ParsePublicKey(register.PubKey))
	response(w, true, util.Encode64(encryptedMsg), u.Token)
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var login model.Credentials
	util.DecodeJSON(req.Body, &login)
	req.Body.Close()

	logger.Info(fmt.Sprintf("Login: %v\n", login))

	u, ok := data.Users[login.User]
	if !ok {
		response(w, false, "Usuario inexistente", nil)
		return
	}

	password := login.Pass

	hash := argon2.Key([]byte(password), u.Salt, 3, 32*1024, 4, 32)
	if !bytes.Equal(u.Hash, hash) {
		response(w, false, "Credenciales inválidas", nil)
	} else {
		u.Seen = time.Now()
		u.Token = make([]byte, 16)
		rand.Read(u.Token)
		data.Users[u.Name] = u
		response(w, true, "Credenciales válidas", u.Token)
	}
}

func getLoginCertHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	username := req.URL.Query().Get("user")

	logger.Info(fmt.Sprintf("Login por certificado GET, %s", username))

	_, ok := data.Users[username]

	if !ok {
		logger.Info(fmt.Sprintf("Usuario %s no encontrado", username))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b := make([]byte, 32)
	rand.Read(b)

	pendingCertLogin[username] = b
	fmt.Fprintf(w, "%s", b)

	go func() {
		// timeout de 5 segundos para que no se llene la memoria de solicitudes
		timer := time.NewTimer(5 * time.Second)
		<-timer.C

		_, ok = pendingCertLogin[username]
		if ok {
			delete(pendingCertLogin, username)
			logger.Info(fmt.Sprintf("Timeout login por certificado para usuario, %s", username))
		}
	}()
}

func postLoginCertHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := req.URL.Query().Get("user")

	logger.Info(fmt.Sprintf("Login por certificado POST, %s", username))

	user, ok := data.Users[username]

	if !ok {
		logger.Info(fmt.Sprintf("Usuario no encontrado, %s", username))
		w.WriteHeader(http.StatusNotFound)
		logger.Info(fmt.Sprintf("Usuario %s no encontrado", username))
		return
	}

	realToken, ok := pendingCertLogin[username]
	if !ok {
		logger.Info("Token expirado")
		w.WriteHeader(http.StatusBadRequest)
		response(w, false, "Token expirado", nil)
		return
	}

	signature := make([]byte, 256)
	req.Body.Read(signature)

	err := util.CheckSignatureRSA(realToken, signature, util.ParsePublicKey(user.PubKey))
	// logger.Info(fmt.Sprintf("checks out %v, real token %v, signature %v, pubkey %v", err == nil, len(realToken), len(signature), len(user.PubKey)))
	// logger.Info(fmt.Sprintf("checks out %v, real token %v, signature %v, pubkey %v", err == nil, realToken, signature, user.PubKey))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Info("Clave incorrecta")
		response(w, false, "Clave incorrecta", nil)
		return
	}

	delete(pendingCertLogin, username)

	user.Token = make([]byte, 16)
	rand.Read(user.Token)

	data.Users[username] = user

	response(w, true, "Autenticación exitosa", []byte(util.Encode64(user.Token)))
}

func postsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var post model.PostContent
	util.DecodeJSON(req.Body, &post)
	req.Body.Close()

	logger.Info(fmt.Sprintf("Creando el post: %v\n", post))

	repository.CreatePost(&data.Posts, &data.UserPosts, &data.GroupPosts, post.Content, req.Header.Get("UserName"), "")

	util.EncodeJSON(model.Resp{Ok: true, Msg: "Post creado", Token: nil})
	response(w, true, "Post creado", nil)
}

func getPostsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(&data.Posts)
	util.FailOnError(err)
}

func chatHandler(w http.ResponseWriter, req *http.Request) {
	logger.Info("Chat handler")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//username := req.Header.Get("UserName")
	otherUser := req.PathValue("user")

	if _, ok := data.Users[otherUser]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// _, ok := data.Chats[NewTupleAlphabeticOrder(username, otherUser)]
	// if !ok {
	// 	data.Chats[NewTupleAlphabeticOrder(username, otherUser)] = make([]model.Message, 0)
	// }

	// test
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Event %d", i)
		logger.Info(fmt.Sprintf("Sending: %s", msg))
		fmt.Fprintf(w, "data: %s\n\n", msg)
		time.Sleep(1 * time.Second)
		w.(http.Flusher).Flush()
	}

	// <-req.Context().Done()
	// fmt.Println("Connection closed")
}

func sendMessageHandler(w http.ResponseWriter, req *http.Request) {
	logger.Info("Message handler")
	w.WriteHeader(http.StatusOK)
}

func getPkHandler(w http.ResponseWriter, req *http.Request) {
	logger.Info("PK handler")
	w.Header().Set("Content-Type", "application/json")

	otherUser := req.PathValue("user")

	u, ok := data.Users[otherUser]
	if !ok {
		response(w, false, "Usuario no encontrado", nil)
		return
	}

	err := json.NewEncoder(w).Encode(u.PubKey)
	util.FailOnError(err)
}

func validarToken(user string, token string) bool {
	u, ok := data.Users[user] // ¿existe ya el usuario?
	if !ok {
		return false
	} else if (u.Token == nil) || (time.Since(u.Seen).Minutes() > 60) {
		return false
	} else if !bytes.EqualFold(u.Token, []byte(token)) {
		return false
	}
	return true
}

func response(w io.Writer, ok bool, msg string, token []byte) {
	r := model.Resp{Ok: ok, Msg: msg, Token: token}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}
