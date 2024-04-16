package handler

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net/http"
	"server/etc"
	"server/logging"
	"strings"
	"time"
	"util"
	"util/model"

	"golang.org/x/crypto/argon2"
)

func RegisterHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var register model.RegisterCredentials

	util.DecodeJSON(req.Body, &register)
	if register.User == "" || register.Pass == "" || register.PubKey == nil {
		etc.Response(w, false, "Campos vacíos", nil)
		return
	}

	if strings.ContainsAny(register.User, "@&?=/:;") {
		etc.Response(w, false, "Carácteres no válidos '@&?=/:;'", nil)
		return
	}

	logMessage := fmt.Sprintf("Registro: %v\n", register)
	logging.SendLogRemote(logMessage)

	w.Header().Set("Content-Type", "application/json")

	data := etc.GetDb(req)

	_, ok := data.Users[register.User]
	if ok {
		etc.Response(w, false, "Usuario ya registrado", nil)
		return
	}

	u := model.User{}
	u.Name = register.User
	u.Salt = make([]byte, 16)
	rand.Read(u.Salt)
	password := register.Pass

	u.Hash = argon2.Key([]byte(password), u.Salt, 3, 32*1024, 4, 32)

	u.Seen = time.Now()
	u.Token = make([]byte, 16)
	rand.Read(u.Token)

	u.PubKey = register.PubKey

	u.Blocked = false
	if len(data.UserNames) == 0 {
		u.Role = model.Admin
	} else {
		u.Role = model.NormalUser
	}

	data.Users[u.Name] = u
	data.UserNames = append(data.UserNames, u.Name)

	encryptedMsg, err := util.EncryptWithRSA([]byte("Bienvenido a la red social"), util.ParsePublicKey(register.PubKey))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		etc.Response(w, false, "Error de clave publica", nil)
		return
	}
	etc.Response(w, true, util.Encode64(encryptedMsg), u.Token)
}

func LoginHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var login model.Credentials
	util.DecodeJSON(req.Body, &login)
	req.Body.Close()

	logMessage := fmt.Sprintf("Login: %v", login)
	logging.SendLogRemote(logMessage)

	data := etc.GetDb(req)

	u, ok := data.Users[login.User]
	if !ok {
		etc.Response(w, false, "Usuario inexistente", nil)
		return
	}

	password := login.Pass

	hash := argon2.Key([]byte(password), u.Salt, 3, 32*1024, 4, 32)
	if !bytes.Equal(u.Hash, hash) {
		etc.Response(w, false, "Credenciales inválidas", nil)
	} else {
		u.Seen = time.Now()
		u.Token = make([]byte, 16)
		rand.Read(u.Token)
		data.Users[u.Name] = u

		logging.SendLogRemote(fmt.Sprintf("Último login del usuario '%s': %s", u.Name, u.Seen.Format(time.RFC3339)))
		etc.Response(w, true, "Credenciales válidas", u.Token)
	}
}

func GetLoginCertHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	username := req.URL.Query().Get("user")

	logging.SendLogRemote(fmt.Sprintf("Login por certificado GET, %s", username))

	data := etc.GetDb(req)

	_, ok := data.Users[username]

	if !ok {
		logging.SendLogRemote(fmt.Sprintf("Usuario %s no encontrado", username))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b := make([]byte, 32)
	rand.Read(b)

	data.PendingCertLogin[username] = b
	fmt.Fprintf(w, "%s", b)

	go func() {
		// timeout de 5 segundos para que no se llene la memoria de solicitudes
		timer := time.NewTimer(5 * time.Second)
		<-timer.C

		_, ok = data.PendingCertLogin[username]
		if ok {
			delete(data.PendingCertLogin, username)
			logging.SendLogRemote(fmt.Sprintf("Timeout login por certificado para usuario, %s", username))
		}
	}()
}

func PostLoginCertHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := req.URL.Query().Get("user")

	logging.SendLogRemote(fmt.Sprintf("Login por certificado POST, %s", username))

	data := etc.GetDb(req)

	user, ok := data.Users[username]

	if !ok {
		logging.SendLogRemote(fmt.Sprintf("Usuario no encontrado, %s", username))
		w.WriteHeader(http.StatusNotFound)
		logging.SendLogRemote(fmt.Sprintf("Usuario %s no encontrado", username))
		return
	}

	realToken, ok := data.PendingCertLogin[username]
	if !ok {
		logging.SendLogRemote("ERROR: Token expirado")
		w.WriteHeader(http.StatusBadRequest)
		etc.Response(w, false, "Token expirado", nil)
		return
	}

	signature := make([]byte, 384)
	req.Body.Read(signature)

	err := util.CheckSignatureRSA(realToken, signature, util.ParsePublicKey(user.PubKey))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logging.SendLogRemote("ERROR: Clave incorrecta")
		etc.Response(w, false, "Clave RSA incorrecta", nil)
		return
	}

	delete(data.PendingCertLogin, username)

	user.Token = make([]byte, 16)
	rand.Read(user.Token)
	user.Seen = time.Now()
	data.Users[username] = user

	logging.SendLogRemote(fmt.Sprintf("Último login del usuario '%s': %s", username, user.Seen.Format(time.RFC3339)))

	etc.Response(w, true, "Autenticación exitosa", []byte(util.Encode64(user.Token)))
}
