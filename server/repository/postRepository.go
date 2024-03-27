package repository

import (
	"strings"
	"time"
	"util/model"
)

var nextIdPosts = 0

func CreatePost(db *model.Database, content string, author string, group string) model.Post {
	post := model.Post{Id: nextIdPosts, Content: strings.TrimSpace(content), Author: author, Group: group, Date: time.Now()}

	// Si post pertenece a grupo, solo sale en feed de grupo, si no, sale publicamente para todos
	if post.Group != "" {
		if _, ok := (*db).GroupPosts[post.Group]; !ok {
			(*db).GroupPosts[post.Group] = make([]int, 2)
		}

		(*db).GroupPosts[post.Group] = append((*db).GroupPosts[post.Group], post.Id)
	} else {
		(*db).Posts[post.Id] = post
		(*db).PostIds = append((*db).PostIds, post.Id)
	}

	(*db).UserPosts[post.Author] = append((*db).UserPosts[post.Author], post.Id)

	nextIdPosts++

	return post
}
