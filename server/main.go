package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
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
		fmt.Printf("Request received from %s\n", req.RemoteAddr)
	})

	fmt.Println("Your server is running on https://localhost:10443")
	err := http.ListenAndServeTLS(":10443", "localhost.crt", "localhost.key", nil)

	if err != nil {
		fmt.Println("Your server is not running on https://localhost:10443")
		println("ListenAndServeTLS:", err.Error())
	}
}
