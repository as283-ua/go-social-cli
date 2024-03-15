package models

import "time"

type User struct {
	Name  string
	Salt  []byte
	Data  map[string]string
	Hash  []byte
	Seen  time.Time
	Token []byte
}

type Post struct {
	Id      int
	Content string
	Author  string
	Date    time.Time
}

type Comments struct {
	Id      int
	Post    int
	Content string
	Author  string
	Date    time.Time
}
