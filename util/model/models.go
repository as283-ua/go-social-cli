package model

import "time"

type User struct {
	Name   string
	Salt   []byte
	Hash   []byte
	Seen   time.Time
	Token  []byte
	PubKey []byte
}

type Group struct {
	Name string
}

type GroupUser struct {
	Group string
	User  string
}

type Post struct {
	Id      int
	Content string
	Author  string
	Group   string
	Date    time.Time
}

type Comments struct {
	Id      int
	Post    int
	Content string
	Author  string
	Date    time.Time
}

type Message struct {
	From string
	Data string
	Read bool
}
