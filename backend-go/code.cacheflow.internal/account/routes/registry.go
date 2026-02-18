package routes

import (
	"net/http"

	"code.cacheflow.internal/account/oauth"
	"code.cacheflow.internal/util/httpx"
	"github.com/go-chi/chi"
)

// RegisterAccountRoutes mounts account-related routes on the provided router.
func RegisterAccountRoutes(r chi.Router) {
	r.Post("/v1/account/create", CreateAccount)
	r.Get("/v1/account/verify", VerifyAccount)
	r.With(oauth.VerifyOAuthToken).Post("/v1/account/verify-token", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, http.StatusOK, map[string]any{
			"message": "Token verified",
		})
	})
	r.With(oauth.VerifyOAuthToken).Get("/v1/account/loadin", LoadInAccountData)
}