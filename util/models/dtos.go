package models

type Resp struct {
	Ok    bool
	Msg   string
	Token []byte
}

type Credentials struct {
	User string
	Pass string
}

type RegisterCredentials struct {
	User   string
	Pass   string
	PubKey []byte
}
