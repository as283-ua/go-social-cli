package repository

import (
	"fmt"
	"time"
	"util/model"
)

func CreateMessage(db *model.Database, sender string, receiver string, message string) {
	key := fmt.Sprintf("%s->%s", sender, receiver)
	if _, ok := db.PendingMessages[key]; !ok {
		db.PendingMessages[key] = make([]model.Message, 0)
	}

	db.PendingMessages[key] = append(db.PendingMessages[key], model.Message{Sender: sender, Message: message, Timestamp: time.Now()})
}

func GetMessages(db *model.Database, sender string, receiver string) []model.Message {
	key := fmt.Sprintf("%s->%s", sender, receiver)

	c := db.PendingMessages[key]

	db.PendingMessages[key] = make([]model.Message, 0)

	return c
}
