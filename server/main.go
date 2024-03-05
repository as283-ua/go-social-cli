package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
	"util"

	"golang.org/x/crypto/argon2"
)

type user struct {
	Name  string            // nombre de usuario
	Hash  []byte            // hash de la contraseña
	Salt  []byte            // sal para la contraseña
	Token []byte            // token de sesión
	Seen  time.Time         // última vez que fue visto
	Data  map[string]string // datos adicionales del usuario
}

var gUsers map[string]user

func chk(e error) {
	if e != nil {
		logger.Error(e.Error())
		panic(e)
	}
}

func getLogger() *slog.Logger {
	return slog.New(slog.Default().Handler())
}

var logger = getLogger()

func main() {
	http.HandleFunc("/", handle)

	logger.Info("Your server is running on https://localhost:10443")
	err := http.ListenAndServeTLS(":10443", "localhost.crt", "localhost.key", nil)

	chk(err)
}

func handle(writer http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	writer.Header().Set("Content-Type", "text/plain")
	chk(err)

	switch req.Form.Get("cmd") { // comprobamos comando desde el cliente
	case "register": // ** registro
		_, ok := gUsers[req.Form.Get("user")] // ¿existe ya el usuario?
		if ok {
			response(writer, false, "Usuario ya registrado", nil)
			return
		}

		u := user{}
		u.Name = req.Form.Get("user")                   // nombre
		u.Salt = make([]byte, 16)                       // sal (16 bytes == 128 bits)
		rand.Read(u.Salt)                               // la sal es aleatoria
		u.Data = make(map[string]string)                // reservamos mapa de datos de usuario
		u.Data["private"] = req.Form.Get("prikey")      // clave privada
		u.Data["public"] = req.Form.Get("pubkey")       // clave pública
		password := util.Decode64(req.Form.Get("pass")) // contraseña (keyLogin)

		// "hasheamos" la contraseña con scrypt (argon2 es mejor)
		u.Hash = argon2.Key(password, u.Salt, 3, 32*1024, 4, 32)

		u.Seen = time.Now()        // asignamos tiempo de login
		u.Token = make([]byte, 16) // token (16 bytes == 128 bits)
		rand.Read(u.Token)         // el token es aleatorio

		gUsers[u.Name] = u
		response(writer, true, "Usuario registrado", u.Token)

	case "login": // ** login
		u, ok := gUsers[req.Form.Get("user")] // ¿existe ya el usuario?
		if !ok {
			response(writer, false, "Usuario inexistente", nil)
			return
		}

		password := util.Decode64(req.Form.Get("pass"))       // obtenemos la contraseña (keyLogin)
		hash := argon2.Key(password, u.Salt, 16384, 8, 1, 32) // scrypt de keyLogin (argon2 es mejor)
		if !bytes.Equal(u.Hash, hash) {                       // comparamos
			response(writer, false, "Credenciales inválidas", nil)

		} else {
			u.Seen = time.Now()        // asignamos tiempo de login
			u.Token = make([]byte, 16) // token (16 bytes == 128 bits)
			rand.Read(u.Token)         // el token es aleatorio
			gUsers[u.Name] = u
			response(writer, true, "Credenciales válidas", u.Token)
		}

	case "data": // ** obtener datos de usuario
		u, ok := gUsers[req.Form.Get("user")] // ¿existe ya el usuario?
		if !ok {
			response(writer, false, "No autentificado", nil)
			return
		} else if (u.Token == nil) || (time.Since(u.Seen).Minutes() > 60) {
			// sin token o con token expirado
			response(writer, false, "No autentificado", nil)
			return
		} else if !bytes.EqualFold(u.Token, util.Decode64(req.Form.Get("token"))) {
			// token no coincide
			response(writer, false, "No autentificado", nil)
			return
		}

		datos, err := json.Marshal(&u.Data) //
		chk(err)
		u.Seen = time.Now()
		gUsers[u.Name] = u
		response(writer, true, string(datos), u.Token)

	default:
		response(writer, false, "Comando no implementado", nil)
	}
}

type Resp struct {
	Ok    bool   // true -> correcto, false -> error
	Msg   string // mensaje adicional
	Token []byte // token de sesión para utilizar por el cliente
}

// función para escribir una respuesta del servidor
func response(w io.Writer, ok bool, msg string, token []byte) {
	r := Resp{Ok: ok, Msg: msg, Token: token} // formateamos respuesta
	rJSON, err := json.Marshal(&r)            // codificamos en JSON
	chk(err)                                  // comprobamos error
	w.Write(rJSON)                            // escribimos el JSON resultante
}
