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

// BD Principal
var Users = make(map[string]model.User)
var Groups = make(map[string]model.Group)
var Posts = make(map[int]model.Post)

/*
 * Se ha sustituido por campo PubKey en model.User
 */
// Se guardan las claves públicas de los usuarios para que puedan iniciar una conversación privada entre ellos
// var UserPubKeys = make(map[string]crypto.PublicKey)

// PK = Post id. No tiene sentido tener una
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

func main() {
	Users = make(map[string]model.User)

	http.HandleFunc("POST /register", registerHandler)
	http.HandleFunc("POST /login", loginHandler)
	http.HandleFunc("GET /users", usersHandler)
	http.HandleFunc("POST /posts", postsHandler)
	http.HandleFunc("GET /posts", getPostsHandler)

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
	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "POST":
		var register model.RegisterCredentials
		util.DecodeJSON[model.RegisterCredentials](req.Body, &register)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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

	if !validarToken(req.Header.Get("UserName"), string(util.Decode64(req.Header.Get("Authorization")))) {
		response(w, false, "Usuario no autenticado", nil)
		return
	}

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

	// if !validarToken(req.Header.Get("UserName"), string(util.Decode64(req.Header.Get("Authorization")))) {
	// 	fmt.Println("Usuario no autenticado")
	// 	response(w, false, "Usuario no autenticado", nil)
	// 	return
	// }

	err := json.NewEncoder(w).Encode(&Posts)
	util.FailOnError(err)
}

// utils server exclusive
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
