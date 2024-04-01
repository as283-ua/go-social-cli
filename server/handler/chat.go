package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/etc"
	"server/logging"
	"time"
	"util"
	"util/model"
)

var ActiveConnections = make(map[string]chan string)
var NewConnections = make(chan string)

func ChatConnectionHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("Chat handler")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//username := req.Header.Get\("Username")
	otherUser := req.PathValue("user")
	// reqUser := req.Header.Get("Username")

	data := etc.GetDb(req)

	if _, ok := data.Users[otherUser]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// otherConn := fmt.Sprintf("%s->%s", otherUser, reqUser)
	// chanReceive, ok := ActiveConnections[otherConn]

	// if !ok {
	// 	closeChan := make(chan bool)
	// 	defer close(closeChan)
	// 	defer closeChan<-true

	// 	go func(){
	// 		select{
	// 		case <-closeChan:
	// 			return
	// 		default:
	// 		}
	// 	}
	// }

	// _, ok := data.Chats[NewTupleAlphabeticOrder(username, otherUser)]
	// if !ok {
	// 	data.Chats[NewTupleAlphabeticOrder(username, otherUser)] = make([]model.Message, 0)
	// }

	// test
	// for i := 0; i < 10; i++ {
	// 	msg := fmt.Sprintf("Event %d", i)
	// 	logging.Info(fmt.Sprintf("Sending: %s", msg))
	// 	fmt.Fprintf(w, "data: %s\n\n", msg)
	// 	time.Sleep(1 * time.Second)
	// 	w.(http.Flusher).Flush()
	// }

	// <-req.Context().Done()
	// fmt.Println("Connection closed")
}

func SendMessageHandler(w http.ResponseWriter, req *http.Request) {
	logging.Info("Send message handler")

	var msg model.Message
	if err := util.DecodeJSON(req.Body, &msg); err != nil {
		logging.Error("Error de JSON")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	otherUser := req.PathValue("user")
	reqUser := req.Header.Get("Username")

	if otherUser == reqUser {
		logging.Error("Nombres iguales")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := etc.GetDb(req)

	_, ok := data.Users[otherUser]

	logging.Info("Hola")

	if !ok {
		logging.Error("Usuario no encontrado")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	key := fmt.Sprintf("%s->%s", reqUser, otherUser)
	messages, ok := data.PendingMessages[key]

	if !ok {
		messages = make([]model.Message, 0)
	}

	msg.Sender = reqUser
	msg.Timestamp = time.Now()

	messages = append(messages, msg)
	data.PendingMessages[key] = messages
}

func GetPendingMessages(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	otherUser := req.PathValue("user")
	reqUser := req.Header.Get("Username")

	data := etc.GetDb(req)

	key := fmt.Sprintf("%s->%s", otherUser, reqUser)

	msgs, ok := data.PendingMessages[key]

	if !ok {
		msgs = make([]model.Message, 0)
	}

	err := json.NewEncoder(w).Encode(msgs)
	if err != nil {
		logging.Error("Error json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	delete(data.PendingMessages, key)
}

func GetPubKeyHandler(w http.ResponseWriter, req *http.Request) {
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
