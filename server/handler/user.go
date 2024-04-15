package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/etc"
	"server/logging"
	"strings"
	"util"
	"util/model"
)

func GetUserNamesHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")

	data := etc.GetDb(req)

	query := req.URL.Query()
	name := query.Get("name")
	usesPagination := query.Get("size") != ""

	var users []model.UserPublicData

	if name == "" {
		n := len(data.UserNames)
		page, size, err := etc.GetPaginationSizes(req, n)

		start, end := etc.PageAndSizeToStartEnd(page, size, n)

		users = make([]model.UserPublicData, end-start)

		if err != nil {
			etc.Response(w, false, "Parametros de paginación incorrectos", nil)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		i := 0
		for _, u := range data.UserNames[start:end] {
			users[i] = model.UserPublicData{Name: u, Blocked: data.Users[u].Blocked, Role: data.Users[u].Role}
			i++
		}
	} else {
		logging.Info(fmt.Sprintf("users with %v", name))

		n := 0x7fffffff

		var (
			page, size, start, end int
		)

		if !usesPagination {
			page = 0
			size = n // por ejemplo
		} else {
			page, size, err = etc.GetPaginationSizes(req, n)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		start, end = etc.PageAndSizeToStartEnd(page, size, n)

		logging.Info(fmt.Sprintf("Desde %v hasta %v", start, end))

		users = make([]model.UserPublicData, 0)

		i := 0
		for _, u := range data.UserNames {
			logging.Info(fmt.Sprintf("%v", u))
			if strings.Contains(u, name) {
				logging.Info(fmt.Sprintf("%v contiene %v", u, name))

				if start <= i && i < end {
					logging.Info(fmt.Sprintf("%v esta entre %v y %v", u, start, end))
					users = append(users, model.MakeUserPublicData(data.Users[u]))
				}

				i++
			}

			if i >= end {
				break
			}
		}
	}

	err = json.NewEncoder(w).Encode(users)
	util.FailOnError(err)
}
