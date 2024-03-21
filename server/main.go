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

var nextIdPosts = 0

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
	http.HandleFunc("/posts", postsHandler)

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
	register := util.DecodeJSON[models.RegisterCredentials](req.Body)
	req.Body.Close()

	// logger.Info(fmt.Sprintf("Registro: %v\n", register))

	w.Header().Set("Content-Type", "application/json")

	_, ok := Users[register.User]
	if ok {
		response(w, false, "Usuario ya registrado", nil)
		UserPubKeys[register.User] = register.PubKey
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

	Users[u.Name] = u
	UserNames = append(UserNames, u.Name)
	UserPubKeys[u.Name] = register.PubKey
	msg := util.EncryptWithRSA([]byte("Bienvenido a la red social"), util.ParsePublicKey(register.PubKey))
	response(w, true, string(msg), u.Token)
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

func postsHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		post := util.DecodeJSON[models.Post](req.Body)
		req.Body.Close()

		post = makePost(post.Content, req.Header.Get("UserName"))

		logger.Info(fmt.Sprintf("Creando el post: %v\n", post))

		if !validarToken(req.Header.Get("UserName"), string(util.Decode64(req.Header.Get("Authorization")))) {
			response(w, false, "Usuario no autenticado", nil)
			return
		}

		posts, ok := UserPosts[req.Header.Get("UserName")]
		if !ok {
			nPosts := make([]models.Post, 0)
			UserPosts[req.Header.Get("UserName")] = nPosts
		}

		posts = append(posts, post)
		UserPosts[req.Header.Get("UserName")] = posts
		util.EncodeJSON(models.Resp{Ok: true, Msg: "Post creado", Token: nil})
		response(w, true, "Post creado", nil)
	case "GET":
		r := UserPosts
		err := json.NewEncoder(w).Encode(&r)
		util.FailOnError(err)
	}

}

func makePost(content string, Author string) models.Post {
	nextIdPosts++
	return models.Post{Content: content, Author: Author, Date: time.Now(), Id: nextIdPosts}
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
	r := models.Resp{Ok: ok, Msg: util.Encode64([]byte(msg)), Token: token}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}
