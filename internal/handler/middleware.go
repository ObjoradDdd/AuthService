package handler

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const UserIDKey contextKey = "userId"

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-User-ID")
		if id == "" {
			http.Error(w, "Unauthorized: missing X-User-ID header", http.StatusUnauthorized)
			return
		}

		userId, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Unauthorized: invalid X-User-ID header", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
