package model

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

type PostContent struct {
	Content string
}

type UserPublicData struct {
	Name    string
	Blocked bool
	Role    Role
}

type Block struct {
	Blocked bool
}

func MakeUserPublicData(user User) UserPublicData {
	return UserPublicData{
		Name:    user.Name,
		Blocked: user.Blocked,
		Role:    user.Role,
	}
}
