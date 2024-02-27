package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", func(http.ResponseWriter, *http.Request) {
		println("Hello, Server!")
	})

	err := http.ListenAndServeTLS(":10443", "localhost.crt", "localhost.key", nil)

	if err != nil {
		println("ListenAndServeTLS:", err.Error())
	}
}
