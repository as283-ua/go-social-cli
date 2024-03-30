package etc

import (
	"encoding/json"
	"io"
	"util"
	"util/model"
)

func Response(w io.Writer, ok bool, msg string, token []byte) {
	r := model.Resp{Ok: ok, Msg: msg, Token: token}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}
