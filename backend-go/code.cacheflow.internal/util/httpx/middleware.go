package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
)

type ctxKey string

const (
	requestIDKey      ctxKey = "request_id"
	RequestIDHeader          = "X-Request-Id"
)

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func RequestID(r *http.Request) string {
	if v := r.Context().Value(requestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(RequestIDHeader)
		if rid == "" {
			rid = newRequestID()
		}
		w.Header().Set(RequestIDHeader, rid)

		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				rid := RequestID(r)
				log.Printf("panic request_id=%s rec=%v", rid, rec)

				status, apiErr := ToAPIError(Internal("Internal server error"), rid)
				WriteJSON(w, status, APIResponse{Ok: false, Error: apiErr})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
