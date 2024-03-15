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
