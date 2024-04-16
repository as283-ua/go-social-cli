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

	logging.SendLogRemote(fmt.Sprintf("Publicar post de %s", req.Header.Get("Username")))

	var postContent model.PostContent
	util.DecodeJSON(req.Body, &postContent)
	req.Body.Close()

	data := etc.GetDb(req)

	post, _ := repository.CreatePost(data, postContent.Content, req.Header.Get("Username"), "")
	logMessage := fmt.Sprintf("Creando el post: %v\n", post)
	logging.SendLogRemote(logMessage)

	etc.Response(w, true, fmt.Sprintf("%v", post.Id), nil)
}

func CreateGroupPostHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupName := req.PathValue("group")

	logging.SendLogRemote(fmt.Sprintf("Publicar post de %s en %s", req.Header.Get("Username"), groupName))

	var postContent model.PostContent
	util.DecodeJSON(req.Body, &postContent)
	req.Body.Close()

	data := etc.GetDb(req)

	post, err := repository.CreatePost(data, postContent.Content, req.Header.Get("Username"), groupName)

	if err != nil {
		logMessage := fmt.Sprintf("Error creando el post %v:%s\n", post, err.Error())
		logging.SendLogRemote(logMessage)
		etc.Response(w, false, logMessage, nil)
		return
	}

	logMessage := fmt.Sprintf("Creando el post: %v\n", post)
	logging.SendLogRemote(logMessage)

	etc.Response(w, true, fmt.Sprintf("%v", post.Id), nil)

}

func GetPostsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logging.SendLogRemote(fmt.Sprintf("Peticion GET para posts en pagina %v", req.URL.Query().Get("page")))

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

	logging.SendLogRemote(fmt.Sprintf("Enviados posts con id: %v", postids))

	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		logging.SendLogRemote("Error enviando")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func GetGroupPostsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logging.SendLogRemote(fmt.Sprintf("Peticion GET para posts en pagina %v", req.URL.Query().Get("page")))

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

	logging.SendLogRemote(fmt.Sprintf("Enviados posts con id: %v", postids))

	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		logging.SendLogRemote("Error enviando")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
