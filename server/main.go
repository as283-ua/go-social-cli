package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/repository"
	"strconv"
	"time"

	"util"
	"util/model"

	"golang.org/x/crypto/argon2"
)

type UserTuple struct {
	First  string
	Second string
}

func NewTupleAlphabeticOrder(a, b string) UserTuple {
	if a <= b {
		return UserTuple{a, b}
	}
	return UserTuple{b, a}
}

// BD Principal
var Users = make(map[string]model.User)
var Groups = make(map[string]model.Group)
var Posts = make(map[int]model.Post)
var Chats = make(map[UserTuple][]model.Message)

/*
 * Se ha sustituido por campo PubKey en model.User
 */
// Se guardan las claves públicas de los usuarios para que puedan iniciar una conversación privada entre ellos
// var UserPubKeys = make(map[string]crypto.PublicKey)

// PK = Post id. No tiene sentido tener una tabla de solo comentarios.
var PostComments = make(map[int][]model.Comments)

// Indexing
// PK = User name
var UserPosts = make(map[string][]int)
var GroupPosts = make(map[string][]int)
var GroupUsers = make(map[string][]string)
var UserGroups = make(map[string][]string)

// extras
var UserNames = make([]string, 0)

var logger = util.GetLogger()

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !validarToken(req.Header.Get("UserName"), string(util.Decode64(req.Header.Get("Authorization")))) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func main() {
	Users = make(map[string]model.User)

	http.HandleFunc("POST /register", registerHandler)
	http.HandleFunc("POST /login", loginHandler)
	http.HandleFunc("GET /users", usersHandler)
	http.Handle("POST /posts", Authorization(http.HandlerFunc(postsHandler)))
	http.HandleFunc("GET /posts", getPostsHandler)
	http.Handle("GET /chat/{user}", Authorization(http.HandlerFunc(chatHandler)))
	http.Handle("GET /noauth/chat/{user}", http.HandlerFunc(chatHandler))
	http.Handle("POST /chat/{user}/message", Authorization(http.HandlerFunc(sendMessageHandler)))
	http.Handle("GET /chat/{user}/pk", Authorization(http.HandlerFunc(getPkHandler)))

	fmt.Printf("Servidor escuchando en https://localhost:10443\n")
	util.FailOnError(http.ListenAndServeTLS(":10443", "localhost.crt", "localhost.key", nil))
}

func usersHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := req.URL.Query()
	pageStr := query.Get("page")
	sizeStr := query.Get("size")
	page := 0
	size := len(UserNames)

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

	if end >= len(UserNames) {
		end = len(UserNames)
	}

	err := json.NewEncoder(w).Encode(UserNames[start:end])
	util.FailOnError(err)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	logger.Info("Register handler")
	w.Header().Set("Content-Type", "application/json")

	var register model.RegisterCredentials
	util.DecodeJSON(req.Body, &register)
	req.Body.Close()

	// logger.Info(fmt.Sprintf("Registro: %v\n", register))

	w.Header().Set("Content-Type", "application/json")

	_, ok := Users[register.User]
	if ok {
		response(w, false, "Usuario ya registrado", nil)
		return
	}

	u := model.User{}
	u.Name = register.User
	u.Salt = make([]byte, 16)
	rand.Read(u.Salt)
	password := register.Pass

	u.Hash = argon2.Key([]byte(password), u.Salt, 16384, 8, 1, 32)

	u.Seen = time.Now()
	u.Token = make([]byte, 16)
	rand.Read(u.Token)

	u.PubKey = register.PubKey
	Users[u.Name] = u
	UserNames = append(UserNames, u.Name)

	msg := util.EncryptWithRSA([]byte("Bienvenido a la red social"), util.ParsePublicKey(register.PubKey))
	response(w, true, string(msg), u.Token)
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var login model.Credentials
	util.DecodeJSON(req.Body, &login)
	req.Body.Close()

	logger.Info(fmt.Sprintf("Login: %v\n", login))

	u, ok := Users[login.User]
	if !ok {
		response(w, false, "Usuario inexistente", nil)
		return
	}

	password := login.Pass
	hash := argon2.Key([]byte(password), u.Salt, 16384, 8, 1, 32)
	if !bytes.Equal(u.Hash, hash) {
		response(w, false, "Credenciales inválidas", nil)
	} else {
		u.Seen = time.Now()
		u.Token = make([]byte, 16)
		rand.Read(u.Token)
		Users[u.Name] = u
		response(w, true, "Credenciales válidas", u.Token)
	}
}

func postsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var post model.PostContent
	util.DecodeJSON(req.Body, &post)
	req.Body.Close()

	logger.Info(fmt.Sprintf("Creando el post: %v\n", post))

	repository.CreatePost(&Posts, &UserPosts, &GroupPosts, post.Content, req.Header.Get("UserName"), "")

	util.EncodeJSON(model.Resp{Ok: true, Msg: "Post creado", Token: nil})
	response(w, true, "Post creado", nil)
}

func getPostsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(&Posts)
	util.FailOnError(err)
}

func chatHandler(w http.ResponseWriter, req *http.Request) {
	logger.Info("Chat handler")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	username := req.Header.Get("UserName")
	otherUser := req.PathValue("user")

	if _, ok := Users[otherUser]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, ok := Chats[NewTupleAlphabeticOrder(username, otherUser)]
	if !ok {
		Chats[NewTupleAlphabeticOrder(username, otherUser)] = make([]model.Message, 0)
	}

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

	u, ok := Users[otherUser]
	if !ok {
		response(w, false, "Usuario no encontrado", nil)
		return
	}

	err := json.NewEncoder(w).Encode(u.PubKey)
	util.FailOnError(err)
}

func validarToken(user string, token string) bool {
	u, ok := Users[user] // ¿existe ya el usuario?
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
	r := model.Resp{Ok: ok, Msg: util.Encode64([]byte(msg)), Token: token}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}
