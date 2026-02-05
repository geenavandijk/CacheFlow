package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/password"
	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OAuthBody struct {
	Username  *string `json:"username"`
	Password  *string `json:"password"`
	GrantType *string `json:"grant_type"`
	Scope     *string `json:"scope"`
}

func OAuthToken(req *http.Request, res http.ResponseWriter, ctx context.Context) *http.Response {

	// Set the logger
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,                  // Report the file name and line number
		ReportTimestamp: true,                  // Report the timestamp
		TimeFormat:      "2006-01-02 15:04:05", // Set the time format
		Prefix:          "OAUTH (OA)",          // Set the prefix
	})

	var body OAuthBody
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		logger.Error("Failed to decode request body")
		return util.JSONResponse(res, http.StatusBadRequest, map[string]interface{}{"error": "invalid request body"})
	}

	if body.Username == nil {
		return util.JSONResponse(res, http.StatusBadRequest, map[string]interface{}{"error": "username is required"})
	}

	username := *body.Username

	username = strings.ToLower(username)

	if body.Scope == nil {
		return util.JSONResponse(res, http.StatusBadRequest, map[string]interface{}{"error": "scope is required"})
	}

	if *body.Scope != "internal" && *body.Scope != "external" {
		return util.JSONResponse(res, http.StatusBadRequest, map[string]interface{}{"error": "invalid scope"})
	}

	accountCollection := datastores.GetMongoDatabase(ctx).Collection(datastores.Accounts)

	accountQuery := bson.D{{Key: "email", Value: username}}
	var account *accountEntities.AccountEntity
	err = accountCollection.FindOne(ctx, accountQuery).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.JSONResponse(res, http.StatusNotFound, map[string]interface{}{"error": "account not found"})
		}
		return util.JSONResponse(res, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	if body.GrantType == nil {
		return util.JSONResponse(res, http.StatusBadRequest, map[string]interface{}{"error": "grant_type is required"})
	}

	grantType := *body.GrantType
	if grantType == "password" {

		verified, err := password.VerifyPassword(body.Password, account.Password.Hash, account.Password.EncryptedVersion)
		if err != nil {
			return util.JSONResponse(res, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}

		if !verified {
			return util.JSONResponse(res, http.StatusUnauthorized, map[string]interface{}{"error": "invalid credentials"})
		}

		if !*account.IsComplete {
			return IssueOAuthResponse(false, &body, req, res, ctx)
		}

		// Account does not have 2FA enabled
		return IssueOAuthResponse(true, &body, req, res, ctx)

	} else if grantType == "refresh" {

		// Get the verified device, check if the refresh token is valid
		sessions := account.Sessions
		var verifiedDevice *accountEntities.Session
		var found bool
		for _, session := range sessions {
			if session.RefreshToken != nil {
				if *session.RefreshToken == *body.Password {
					verifiedDevice = session
					found = true
					break
				}
			}
		}

		if !found {
			return util.JSONResponse(res, http.StatusUnauthorized, map[string]interface{}{"error": "invalid refresh token"})
		}

		// Check if the refresh token is expired
		if verifiedDevice.RefreshIssuedAt.Add(time.Hour * 24 * 28).Before(time.Now()) {
			return util.JSONResponse(res, http.StatusUnauthorized, map[string]interface{}{"error": "refresh token expired"})
		}

		accessToken, refreshToken, _, _, _, err := RefreshOAuth2Token("external", req, res, body.Password)
		if err != nil {
			return util.JSONResponse(res, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}

		return util.JSONResponse(res, http.StatusOK, map[string]interface{}{
			"access_token":  *accessToken,
			"token_type":    "bearer",
			"refresh_token": *refreshToken,
		})
	} else if *body.GrantType == "revoke" {

		err := RevokeOAuth2Token(req, res, ctx)
		if err != nil {
			return util.JSONResponse(res, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}

		return util.JSONResponse(res, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "session revoked",
		})

	} else {
		return util.JSONResponse(res, http.StatusBadRequest, map[string]interface{}{"error": "invalid grant_type"})
	}
}

func IssueOAuthResponse(accountComplete bool, body *OAuthBody, req *http.Request, res http.ResponseWriter, ctx context.Context) *http.Response {

	// Issue token
	token, _, refreshToken, _, _, err := GenerateOAuth2Token(req, res, ctx)
	if err != nil {
		return util.JSONResponse(res, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	redirectToOnboarding := !accountComplete

	return util.JSONResponse(res, http.StatusOK, map[string]interface{}{
		"access_token":        *token,
		"token_type":          "bearer",
		"refresh_token":       *refreshToken,
		"onboarding_redirect": redirectToOnboarding,
		"success":             true,
	})
}
