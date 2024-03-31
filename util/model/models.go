package model

import "time"

// BD Principal
type Database struct {
	Users            map[string]User
	Groups           map[string]Group
	Posts            map[int]Post
	UserPosts        map[string][]int
	GroupPosts       map[string][]int
	GroupUsers       map[string][]string
	UserGroups       map[string][]string
	UserNames        []string
	PostIds          []int
	NextPostId       int
	PendingCertLogin map[string][]byte
	PendingMessages  map[string][]Message
}

/*
Pending Chat Messages: la clave ser aun string con formato usuario1->usuario2. La clave es una lista de mensajes que el usuario1
ha enviado al usuario2, que el usuario 2 aun no ha leido. Al recibir dichos mensajes (solo descifrables por el usuario2) se borran de
esta tabla.
*/

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

type Message struct {
	Sender    string
	Message   string
	Timestamp time.Time
}

type Chat struct {
	UserA    string
	UserB    string
	Messages []Message
}
