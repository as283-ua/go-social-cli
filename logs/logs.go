package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"util"
	"util/model"
)

const key = "clave_secreta"

func main() {

	file, err := os.OpenFile("logs.log", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create("logs.log")
			util.FailOnError(err)
		} else {
			util.FailOnError(err)
		}
	}

	defer file.Close()

	server := http.Server{
		Addr: ":10444",
	}

	router := http.NewServeMux()

	router.HandleFunc("POST /logs", func(w http.ResponseWriter, r *http.Request) { logsHandler(w, r, file) })

	server.Handler = router
	fmt.Printf("Servidor escuchando en https://localhost:10444\n")
	util.FailOnError(server.ListenAndServeTLS("localhost.crt", "localhost.key"))
}

func logsHandler(w http.ResponseWriter, req *http.Request, file *os.File) {

	w.Header().Set("Content-Type", "text/plain")

	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authKey := strings.TrimSpace(authHeader)
	if authKey != key {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var log string
	body, err := io.ReadAll(req.Body)
	util.FailOnError(err)

	log = string(body)
	req.Body.Close()

	writer := bufio.NewWriter(file)

	_, err = writer.WriteString("\n" + log)
	util.FailOnError(err)

	err = writer.Flush()
	util.FailOnError(err)

	fmt.Println(log)

	response(w, true, "Log creado")
}

func response(w io.Writer, ok bool, msg string) {
	r := model.Resp{Ok: ok, Msg: util.Encode64([]byte(msg))}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}
