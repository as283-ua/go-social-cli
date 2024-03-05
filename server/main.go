package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type user struct {
	Name  string            // nombre de usuario
	Hash  []byte            // hash de la contraseña
	Salt  []byte            // sal para la contraseña
	Token []byte            // token de sesión
	Seen  time.Time         // última vez que fue visto
	Data  map[string]string // datos adicionales del usuario
}

var gUsers map[string]user

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func getLogger() *slog.Logger {
	return slog.New(slog.Default().Handler())
}

var logger = getLogger()

func main() {
	http.HandleFunc("/", handle)

	logger.Info("Your server is running on https://localhost:10443")
	err := http.ListenAndServeTLS(":10443", "localhost.crt", "localhost.key", nil)

	chk(err)
}

func handle(writer http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		logger.Error("Error parsing form data:", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	writer.Write([]byte(`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Document</title>
	</head>
	<body>
		<h1>Hello, Server!</h1>
	</body>
	</html>`))
	bodyStr, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("Error reading body:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(string(bodyStr))
	logger.Info("Header", "header", req.Header)
	logger.Info("Request", "method", req.Method, "url", req.URL)
	logger.Info("Form", "form", req.Form)
	fmt.Println()
}
