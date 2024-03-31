package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/etc"
	"server/logging"
	"time"
	"util"
)

func ChatHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("Chat handler")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//username := req.Header.Get\("Username")
	otherUser := req.PathValue("user")

	data := etc.GetDb(req)

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

func SendMessageHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("Message handler")
	w.WriteHeader(http.StatusOK)
}

func GetPubKeyHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("PK handler")
	w.Header().Set("Content-Type", "application/json")

	otherUser := req.PathValue("user")

	data := etc.GetDb(req)

	u, ok := data.Users[otherUser]
	if !ok {
		etc.Response(w, false, "Usuario no encontrado", nil)
		return
	}

	err := json.NewEncoder(w).Encode(u.PubKey)
	util.FailOnError(err)
}
