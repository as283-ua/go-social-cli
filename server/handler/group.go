package handler

import (
	"fmt"
	"net/http"
	"server/etc"
	"server/logging"
	"server/repository"
	"util"
	"util/model"
)

func CreateGroupHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var group model.Group
	util.DecodeJSON(req.Body, &group)
	req.Body.Close()

	logging.SendLogRemote(fmt.Sprintf("Crear grupo %s:  %s", group.Name, req.Header.Get("Username")))

	data := etc.GetDb(req)

	group, err := repository.CreateGroup(data, group.Name)

	if err != nil {
		logging.SendLogRemote(err.Error())
		etc.ResponseSimple(w, false, fmt.Sprintf("%v", err.Error()))
	} else {
		data.GroupUsers[group.Name] = append(data.GroupUsers[group.Name], req.Header.Get("Username"))

		logging.SendLogRemote(fmt.Sprintf("Grupo creado: %s\n", group))
		etc.ResponseSimple(w, true, fmt.Sprintf("%v", group.Name))
	}

}

func JoinGroupHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupName := req.PathValue("group")

	logging.SendLogRemote(fmt.Sprintf("Unirse al grupo %s:  %s", groupName, req.Header.Get("Username")))

	data := etc.GetDb(req)

	if _, existe := data.Groups[groupName]; existe {
		if repository.JoinGroup(data, groupName, req.Header.Get("Username")) {
			logging.SendLogRemote(fmt.Sprintf("Agregado al grupo  %s:  %s", groupName, req.Header.Get("Username")))
			etc.ResponseSimple(w, true, "Agregado al grupo")
		} else {
			logging.SendLogRemote("El usuario ya es miembro")
			etc.ResponseSimple(w, false, "El usuario ya es miembro")
		}
	} else {
		logging.SendLogRemote("El grupo no existe")
		etc.ResponseSimple(w, false, "El grupo no existe")
	}
}

func UserCanAccessGroupHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupName := req.PathValue("group")

	logging.SendLogRemote(fmt.Sprintf("Comprobando acceso al grupo %s:  %s", groupName, req.Header.Get("Username")))

	data := etc.GetDb(req)

	if repository.UserCanAccessGroup(data, groupName, req.Header.Get("Username")) {
		logging.SendLogRemote(fmt.Sprintf("Usuario %s tiene acceso al grupo %s", groupName, req.Header.Get("Username")))
		etc.ResponseSimple(w, true, "Acceso permitido")
	} else {
		logging.SendLogRemote(fmt.Sprintf("Usuario %s no tiene acceso al grupo %s", groupName, req.Header.Get("Username")))
		etc.ResponseSimple(w, false, "Acceso denegado")
	}
}
