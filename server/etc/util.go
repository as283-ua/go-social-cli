package etc

import (
	"encoding/json"
	"io"
	"net/http"
	"server/middleware"
	"strconv"
	"util"
	"util/model"
)

func ResponseSimple(w io.Writer, ok bool, msg string) {
	r := model.Resp{Ok: ok, Msg: msg}
	err := json.NewEncoder(w).Encode(&r)
	util.FailOnError(err)
}

func ResponseAuth(w io.Writer, ok bool, msg string, user model.User) {
	r := model.RespAuth{Ok: ok, Msg: msg, User: user}
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

func GetPaginationSizes(req *http.Request, dataLength int) (int, int, error) {

	query := req.URL.Query()
	pageStr := query.Get("page")
	sizeStr := query.Get("size")
	page := 0
	size := dataLength

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, err
		}
		page = p
	}

	if sizeStr != "" {
		s, err := strconv.Atoi(sizeStr)
		if err != nil {
			return 0, 0, err
		}
		size = s
	}

	return page, size, nil
}

func PageAndSizeToStartEnd(page, size, dataLength int) (start, end int) {
	start = page * size
	end = (page + 1) * size

	if end >= dataLength {
		end = dataLength
	}

	if start >= dataLength {
		start = dataLength
	}

	if end-start < 0 {
		end = start
	}

	return
}
