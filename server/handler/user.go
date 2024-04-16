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
		logging.SendLogRemote(fmt.Sprintf("Users with %v", name))

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

		logging.SendLogRemote(fmt.Sprintf("Desde %v hasta %v", start, end))

		users = make([]model.UserPublicData, 0)

		i := 0
		for _, u := range data.UserNames {
			logging.SendLogRemote(fmt.Sprintf("%v", u))

			if strings.Contains(u, name) {
				logging.SendLogRemote(fmt.Sprintf("%v contiene %v", u, name))

				if start <= i && i < end {
					logging.SendLogRemote(fmt.Sprintf("%v esta entre %v y %v", u, start, end))
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

func SetBlocked(w http.ResponseWriter, req *http.Request) {
	otherUser := req.PathValue("user")

	data := etc.GetDb(req)

	if u, ok := data.Users[otherUser]; ok {
		var block model.Block
		err := util.DecodeJSON(req.Body, &block)

		if err != nil {
			etc.Response(w, false, "Error interno", nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		u.Blocked = block.Blocked

		data.Users[otherUser] = u

		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
