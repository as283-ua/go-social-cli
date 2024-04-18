package repository

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"util/model"
)

func CreatePost(db *model.Database, content string, author string, group string) (model.Post, error) {
	if strings.TrimSpace(content) == "" {
		return model.Post{}, fmt.Errorf("no puedes publicar un post vacío")
	}
	post := model.Post{Id: db.NextPostId, Content: strings.TrimSpace(content), Author: author, Group: group, Date: time.Now()}

	// Si post pertenece a grupo, solo sale en feed de grupo, si no, sale publicamente para todos
	if post.Group != "" {
		if !UserCanAccessGroup(db, group, author) {
			return post, fmt.Errorf("el usuario no tiene acceso al grupo")
		}

		(*db).GroupPosts[post.Id] = post
		(*db).GroupPostIds[post.Group] = append((*db).GroupPostIds[post.Group], post.Id)
	} else {
		(*db).Posts[post.Id] = post
		newPost := make([]int, 1)
		newPost[0] = post.Id
		(*db).PostIds = slices.Concat(newPost, (*db).PostIds)
	}

	(*db).UserPosts[post.Author] = append((*db).UserPosts[post.Author], post.Id)

	db.NextPostId++

	return post, nil
}
