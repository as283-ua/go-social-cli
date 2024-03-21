package repository

import (
	"time"
	"util/models"
)

var nextIdPosts = 0

func CreatePost(posts *map[int]models.Post, userPosts *map[string][]int, groupPosts *map[string][]int, content string, author string, group string) {
	post := models.Post{Id: nextIdPosts, Content: content, Author: author, Group: group, Date: time.Now()}

	// Si post pertenece a grupo, solo sale en feed de grupo, si no, sale publicamente para todos
	if post.Group != "" && groupPosts != nil {
		(*groupPosts)[post.Group] = append((*groupPosts)[post.Group], post.Id)
	} else {
		(*posts)[post.Id] = post
	}

	(*userPosts)[post.Author] = append((*userPosts)[post.Author], post.Id)

	nextIdPosts++
}
