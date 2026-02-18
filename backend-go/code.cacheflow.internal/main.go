package main

import (
	"net/http"
	"os"

	"code.cacheflow.internal/account/oauth"
	accountRoutes "code.cacheflow.internal/account/routes"
	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/httpx"
	"code.cacheflow.internal/util/secrets"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Prefix:          "SERVER",
	})

	logger.Info("Starting CacheFlow Internal Server")

	secrets.InitializeSecretCache()
	logger.Info("Secrets initialized")

	datastores.ConnectDB(secrets.DatabaseSecretValue)

	r := chi.NewRouter()

	// ✅ Centralized error handling base
	r.Use(httpx.WithRequestID)
	r.Use(httpx.Recover)

	// keep existing JSON middleware
	r.Use(util.JSONMiddleware)

	// CORS (includes your frontend headers + request id)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type", "X-CSRF-Token",
			"x-cf-device-id", "x-cf-uid", "x-cf-bearer", "x-cf-refresh",
			"X-Request-Id", "x-cf-auth-scope",
		},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// JSON 404 / 405
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, httpx.NotFound("Route not found"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, httpx.BadRequest("Method not allowed", map[string]string{
			"method": "not allowed",
		}))
	})

	// Account routes
	accountRoutes.RegisterAccountRoutes(r)

	// Health route (standard response)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, http.StatusOK, map[string]any{
			"message": "CacheFlow Internal Server is running",
			"version": "1.0.0",
		})
	})

	r.Post("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		oauth.OAuthToken(r, w, r.Context())
	})

	logger.Info("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Error(err.Error())
	}
}
