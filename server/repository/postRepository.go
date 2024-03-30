package repository

import (
	"slices"
	"strings"
	"time"
	"util/model"
)

func CreatePost(db *model.Database, content string, author string, group string) model.Post {
	post := model.Post{Id: db.NextPostId, Content: strings.TrimSpace(content), Author: author, Group: group, Date: time.Now()}

	// Si post pertenece a grupo, solo sale en feed de grupo, si no, sale publicamente para todos
	if post.Group != "" {
		if _, ok := (*db).GroupPosts[post.Group]; !ok {
			(*db).GroupPosts[post.Group] = make([]int, 2)
		}

		(*db).GroupPosts[post.Group] = append((*db).GroupPosts[post.Group], post.Id)
	} else {
		(*db).Posts[post.Id] = post
		newPost := make([]int, 1)
		newPost[0] = post.Id
		(*db).PostIds = slices.Concat(newPost, (*db).PostIds)
	}

	(*db).UserPosts[post.Author] = append((*db).UserPosts[post.Author], post.Id)

	db.NextPostId++

	return post
}
