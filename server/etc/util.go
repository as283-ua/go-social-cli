package etc

import (
	"encoding/json"
	"io"
	"net/http"
	"server/middleware"
	"util"
	"util/model"
)

func Response(w io.Writer, ok bool, msg string, token []byte) {
	r := model.Resp{Ok: ok, Msg: msg, Token: token}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}

func GetDb(req *http.Request) *model.Database {
	db := req.Context().Value(middleware.ContextKeyData)
	if db == nil {
		return nil
	}
	return db.(*model.Database)
}
