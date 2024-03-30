package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"server/etc"
	"server/logging"
	"server/middleware"
	"server/repository"
	"strconv"
	"strings"
	"syscall"
	"time"
	"util"
	"util/model"
)

func NewTupleAlphabeticOrder(a, b string) model.UserChat {
	if a <= b {
		return model.UserChat{First: a, Second: b}
	}
	return model.UserChat{First: b, Second: a}
}

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

	server := http.Server{
		Addr: ":10443",
	}

	router := http.NewServeMux()

	router.HandleFunc("POST /register", handler.registerHandler)
	router.HandleFunc("POST /login", loginHandler)
	router.HandleFunc("GET /login/cert", getLoginCertHandler)
	router.HandleFunc("POST /login/cert", postLoginCertHandler)
	router.HandleFunc("GET /users", usersHandler)
	router.Handle("POST /posts", middleware.Authorization(http.HandlerFunc(postsHandler), &data))
	router.HandleFunc("GET /posts", getPostsHandler)
	router.Handle("GET /chat/{user}", middleware.Authorization(http.HandlerFunc(chatHandler), &data))
	router.Handle("GET /noauth/chat/{user}", http.HandlerFunc(chatHandler))
	router.Handle("POST /chat/{user}/message", middleware.Authorization(http.HandlerFunc(sendMessageHandler), &data))
	router.Handle("GET /chat/{user}/pk", middleware.Authorization(http.HandlerFunc(getPkHandler), &data))

	server.Handler = router
	fmt.Printf("Servidor escuchando en https://localhost:10443\n")
	util.FailOnError(server.ListenAndServeTLS("localhost.crt", "localhost.key"))
}

func getPaginationSizes(req *http.Request) (int, int, error) {

	query := req.URL.Query()
	pageStr := query.Get("page")
	sizeStr := query.Get("size")
	page := 0
	size := len(data.UserNames)

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, err
		}
		page = p
	}

	if sizeStr != "" {
		s, err := strconv.Atoi(sizeStr)
		if err != nil {
			return 0, 0, err
		}
		size = s
	}

	return page, size, nil
}

func usersHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page, size, err := getPaginationSizes(req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	start := page * size
	end := (page + 1) * size

	if end >= len(data.UserNames) {
		end = len(data.UserNames)
	}

	err = json.NewEncoder(w).Encode(data.UserNames[start:end])
	util.FailOnError(err)
}

func postsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logging.Info(fmt.Sprintf("Publicar post de %s", req.Header.Get("Username")))

	var postContent model.PostContent
	util.DecodeJSON(req.Body, &postContent)
	req.Body.Close()

	post := repository.CreatePost(&data, postContent.Content, req.Header.Get("Username"), "")
	logMessage := fmt.Sprintf("Creando el post: %v\n", post)
	logging.Info(logMessage)
	logging.SendLogRemote(logMessage)

	util.EncodeJSON(model.Resp{Ok: true, Msg: fmt.Sprintf("%v", post.Id), Token: nil})
	etc.Response(w, true, "Post creado", nil)
}

func getPostsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logging.Info(fmt.Sprintf("Peticion GET para posts en pagina %v", req.URL.Query().Get("page")))

	page, size, err := getPaginationSizes(req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	start := page * size
	end := start + size
	n := len(data.PostIds)
	if end > n {
		end = n
	}

	var postids []int
	if n <= start {
		postids = nil
		end = 0
		start = 0
	} else {
		if n < end {
			end = n
		}
		postids = data.PostIds[start:end]
	}

	posts := make([]model.Post, end-start)
	for i, id := range postids {
		posts[i] = data.Posts[id]
	}

	logging.Info(fmt.Sprintf("Enviados posts con id: %v", postids))

	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		logging.Error("Error enviando")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func chatHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("Chat handler")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//username := req.Header.Get\("Username")
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
		logging.Info(fmt.Sprintf("Sending: %s", msg))
		fmt.Fprintf(w, "data: %s\n\n", msg)
		time.Sleep(1 * time.Second)
		w.(http.Flusher).Flush()
	}

	// <-req.Context().Done()
	// fmt.Println("Connection closed")
}

func sendMessageHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("Message handler")
	w.WriteHeader(http.StatusOK)
}

func getPkHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("PK handler")
	w.Header().Set("Content-Type", "application/json")

	otherUser := req.PathValue("user")

	u, ok := data.Users[otherUser]
	if !ok {
		etc.Response(w, false, "Usuario no encontrado", nil)
		return
	}

	err := json.NewEncoder(w).Encode(u.PubKey)
	util.FailOnError(err)
}
