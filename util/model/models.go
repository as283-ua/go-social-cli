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
}

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
	From string
	Data string
	Read bool
}

type ChatMessage struct {
	Sender  string
	Message string
}

type Chat struct {
	UserA    string
	UserB    string
	Messages []ChatMessage
}
