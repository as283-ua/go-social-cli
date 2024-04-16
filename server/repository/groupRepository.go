package repository

import (
	"fmt"
	"slices"
	"util/model"
)

func CreateGroup(db *model.Database, name string) (model.Group, error) {
	group := model.Group{Name: name}

	if _, existe := db.Groups[name]; existe {
		return group, fmt.Errorf("el grupo ya existe")
	} else {
		(*db).Groups[name] = group
	}

	return group, nil
}

func JoinGroup(db *model.Database, group string, user string) bool {

	if (*db).GroupUsers[group] == nil {
		(*db).GroupUsers[group] = make([]string, 0)
	}

	if slices.Contains((*db).GroupUsers[group], user) {
		return false
	} else {
		(*db).GroupUsers[group] = append((*db).GroupUsers[group], user)
	}

	return true
}

func UserCanAccessGroup(db *model.Database, group string, user string) bool {
	if (*db).GroupUsers[group] == nil {
		return false
	}

	return slices.Contains((*db).GroupUsers[group], user)
}
