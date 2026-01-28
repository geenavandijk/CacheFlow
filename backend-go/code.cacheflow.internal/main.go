package main

import (
	"net/http"
	"os"

	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/secrets"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"

	datastores "code.cacheflow.internal/datastores/mongo"
)

func main() {

	// Initialize logger
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,                  // Report the file name and line number
		ReportTimestamp: true,                  // Report the timestamp
		TimeFormat:      "2006-01-02 15:04:05", // Set the time format
		Prefix:          "SERVER",              // Set the prefix
	})

	logger.Info("Starting CacheFlow Internal Server")

	// Initialize secrets
	secrets.InitializeSecretCache()

	logger.Info("Secrets initialized")

	datastores.ConnectDB(secrets.DatabaseSecretValue)

	r := chi.NewRouter()

	r.Use(util.JSONMiddleware)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"}, // Allow your frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-CacheFlow-Username", "X-CacheFlow-Device-Id", "*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		util.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"message": "CacheFlow Internal Server is running",
			"version": "1.0.0",
		})
	})

	logger.Info("Server started at http://localhost:8080")
	logger.Error(http.ListenAndServe(":8080", r).Error())
}