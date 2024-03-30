package middleware

import (
	"context"
	"net/http"
	"util/model"
)

type contextKey string

const (
	ContextKeyData = contextKey("db")
)

func InjectData(data *model.Database) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), ContextKeyData, data)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}
