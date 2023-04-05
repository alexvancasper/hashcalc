package middleware

import (
	"context"
	"net/http"

	guid "github.com/satori/go.uuid"
)

const (
	RequestContextID = "requestID"
)

func UUIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = guid.Must(guid.NewV4(), nil).String()
		}
		ctx = context.WithValue(r.Context(), RequestContextID, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
