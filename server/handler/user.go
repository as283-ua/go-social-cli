package handler

import (
	"encoding/json"
	"net/http"
	"server/etc"
	"util"
)

func GetUserNamesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := etc.GetDb(req)

	page, size, err := etc.GetPaginationSizes(req, len(data.UserNames))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	start := page * size
	end := (page + 1) * size

	if end >= len(data.UserNames) {
		end = len(data.UserNames)
	}

	err = json.NewEncoder(w).Encode(data.UserNames[start:end])
	util.FailOnError(err)
}
