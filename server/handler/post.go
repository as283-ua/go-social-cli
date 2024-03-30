package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/etc"
	"server/logging"
	"server/repository"
	"util"
	"util/model"
)

func CreatePostHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logging.Info(fmt.Sprintf("Publicar post de %s", req.Header.Get("Username")))

	var postContent model.PostContent
	util.DecodeJSON(req.Body, &postContent)
	req.Body.Close()

	data := etc.GetDb(req)

	post := repository.CreatePost(data, postContent.Content, req.Header.Get("Username"), "")
	logMessage := fmt.Sprintf("Creando el post: %v\n", post)
	logging.Info(logMessage)
	logging.SendLogRemote(logMessage)

	util.EncodeJSON(model.Resp{Ok: true, Msg: fmt.Sprintf("%v", post.Id), Token: nil})
	etc.Response(w, true, "Post creado", nil)
}

func GetPostsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logging.Info(fmt.Sprintf("Peticion GET para posts en pagina %v", req.URL.Query().Get("page")))

	data := etc.GetDb(req)

	page, size, err := etc.GetPaginationSizes(req, len(data.PostIds))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	start := page * size
	end := start + size
	n := len(data.PostIds)
	if end > n {
		end = n
	}

	var postids []int
	if n <= start {
		postids = nil
		end = 0
		start = 0
	} else {
		if n < end {
			end = n
		}
		postids = data.PostIds[start:end]
	}

	posts := make([]model.Post, end-start)
	for i, id := range postids {
		posts[i] = data.Posts[id]
	}

	logging.Info(fmt.Sprintf("Enviados posts con id: %v", postids))

	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		logging.Error("Error enviando")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
