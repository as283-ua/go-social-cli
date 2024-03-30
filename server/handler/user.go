package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/etc"
	"server/logging"
	"strconv"
	"strings"
	"util"
)

func GetUserNamesHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")

	data := etc.GetDb(req)

	query := req.URL.Query()
	name := query.Get("name")
	usesPagination := query.Get("size") != ""

	var users []string

	if name == "" {
		users = data.UserNames
	} else {
		logging.Info(fmt.Sprintf("users with %v", name))
		var count int
		if usesPagination {
			count, err = strconv.Atoi(query.Get("size"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			count = 10 // por ejemplo
		}

		users = make([]string, 0, count)

		for _, u := range data.UserNames {
			if strings.Contains(u, name) {
				users = append(users, u)
			}
		}
	}

	page, size, err := etc.GetPaginationSizes(req, len(users))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	start := page * size
	end := (page + 1) * size

	if end >= len(users) {
		end = len(users)
	}

	logging.Info(fmt.Sprintf("GET users, nombre conteniendo '%s', pagina %s, tamaño %s", name, query.Get("page"), query.Get("size")))

	if start > end {
		err = json.NewEncoder(w).Encode(make([]string, 0))
		util.FailOnError(err)
		return
	}

	err = json.NewEncoder(w).Encode(users[start:end])
	util.FailOnError(err)
}
