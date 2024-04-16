package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"server/logging"
	"time"
	"util"
	"util/model"
)

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token, err := util.Decode64(req.Header.Get("Authorization"))

		// logging.Info(fmt.Sprintf("Token %v", token))
		if err != nil {
			logging.Info("error de login. No se ha podido decodificar el header 'Authorization'")
			w.WriteHeader(http.StatusInternalServerError)
			util.FailOnError(err)
			return
		}

		data := req.Context().Value(ContextKeyData).(*model.Database)

		if data == nil {
			logging.Info("DB nil")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		username := req.Header.Get("Username")

		if err := validarToken(username, token, data); err != nil {
			logging.Info(fmt.Sprintf("error de login. %s", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if data.Users[username].Blocked {
			logging.Info(fmt.Sprintf("error de login. '%s' esta bloqueado", username))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		username := req.Header.Get("Username")

		data := req.Context().Value(ContextKeyData).(*model.Database)

		user := data.Users[username]

		if user.Role != model.Admin {
			logging.Info(fmt.Sprintf("error de autorización. Usuario '%s' no es admin", username))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func validarToken(user string, token []byte, data *model.Database) error {
	if user == "" {
		return fmt.Errorf("nombre de usuario no proporcionado")
	}

	if token == nil {
		return fmt.Errorf("token no proporcionado")
	}

	u, ok := data.Users[user] // ¿existe ya el usuario?
	if !ok {
		return fmt.Errorf("usuario no encontrado")
	} else if time.Since(u.Seen).Minutes() > 60 {
		return fmt.Errorf("token expirado")
	} else if !bytes.EqualFold(u.Token, token) {
		return fmt.Errorf(fmt.Sprintf("token incorrecto. Real: %v. Proporcionado: %v", u.Token, token))
	}

	return nil
}
