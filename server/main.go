package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"util"
	"util/models"

	"golang.org/x/crypto/argon2"
)

var UserNames = make([]string, 0)
var Users = make(map[string]models.User)

// Se guardan las claves públicas de los usuarios para que puedan iniciar una conversación privada entre ellos
var UserPubKeys = make(map[string]crypto.PublicKey)

// PK = User name
var UserPosts = make(map[string][]models.Post)

// PK = Post id
var PostComments = make(map[int][]models.Comments)

// var UserFollowers map[string][]string
// var UserFollowing map[string][]string

var logger = util.GetLogger()

func main() {
	Users = make(map[string]models.User)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/users", usersHandler)

	fmt.Printf("Servidor escuchando en https://localhost:10443\n")
	util.FailOnError(http.ListenAndServeTLS(":10443", "localhost.crt", "localhost.key", nil))
}

func usersHandler(w http.ResponseWriter, req *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")

	start := page * size
	end := (page + 1) * size

	if end >= len(UserNames) {
		end = len(UserNames)
	}

	err := json.NewEncoder(w).Encode(UserNames[start:end])
	util.FailOnError(err)
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	register := util.DecodeJSON[models.Credentials](req.Body)
	req.Body.Close()

	logger.Info(fmt.Sprintf("Registro: %v\n", register))

	w.Header().Set("Content-Type", "application/json")

	_, ok := Users[register.User]
	if ok {
		response(w, false, "Usuario ya registrado", nil)
		return
	}

	u := models.User{}
	u.Name = register.User
	u.Salt = make([]byte, 16)
	rand.Read(u.Salt)
	u.Data = make(map[string]string)
	password := register.Pass

	u.Hash = argon2.Key([]byte(password), u.Salt, 16384, 8, 1, 32)

	u.Seen = time.Now()
	u.Token = make([]byte, 16)
	rand.Read(u.Token)

	// logger.Info(util.Encode64(u.Token))

	Users[u.Name] = u
	UserNames = append(UserNames, u.Name)
	response(w, true, "Usuario registrado", u.Token)
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	login := util.DecodeJSON[models.Credentials](req.Body)
	req.Body.Close()

	logger.Info(fmt.Sprintf("Login: %v\n", login))

	w.Header().Set("Content-Type", "application/json")

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

func response(w io.Writer, ok bool, msg string, token []byte) {
	r := models.Resp{Ok: ok, Msg: msg, Token: token}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}
