package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/util"
	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

type UpdateSettingsBody struct {
	FirstName    *string               `json:"first_name"`
	LastName     *string               `json:"last_name"`
	RiskSettings *UpdateRiskSettingsBody `json:"risk_settings"`
}

type UpdateRiskSettingsBody struct {
	Budget            *int64 `json:"budget"`
	MaxLossPercentage *int64 `json:"max_loss_percentage"`
	RiskTolerance     *int64 `json:"risk_tolerance"`
}

func UpdateAccountSettings(res http.ResponseWriter, req *http.Request) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Prefix:          "ACCOUNT (US)",
	})

	if req.Method != http.MethodPatch && req.Method != http.MethodPut {
		util.JSONResponse(res, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
		return
	}

	email := req.Header.Get("x-cf-uid")
	if email == "" {
		util.JSONResponse(res, http.StatusUnauthorized, map[string]any{
			"error": "email is required",
		})
		return
	}
	email = strings.ToLower(strings.TrimSpace(email))

	var body UpdateSettingsBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Error("failed to decode request body", "err", err)
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "invalid request body",
		})
		return
	}

	ctx := req.Context()
	db := datastores.GetMongoDatabase(ctx)
	accountsCollection := db.Collection(datastores.Accounts)

	var existing accountEntities.AccountEntity
	err := accountsCollection.FindOne(ctx, bson.M{"email": email}).Decode(&existing)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			util.JSONResponse(res, http.StatusNotFound, map[string]any{
				"error": "account not found",
			})
			return
		}
		logger.Error("failed to find account", "err", err)
		util.JSONResponse(res, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}

	// Build update document: only set fields that were provided
	setFields := bson.D{{Key: "updated_at", Value: time.Now()}}

	if body.FirstName != nil {
		setFields = append(setFields, bson.E{Key: "first_name", Value: strings.TrimSpace(*body.FirstName)})
	}
	if body.LastName != nil {
		setFields = append(setFields, bson.E{Key: "last_name", Value: strings.TrimSpace(*body.LastName)})
	}

	if body.RiskSettings != nil {
		rs := body.RiskSettings
		if rs.Budget != nil && *rs.Budget >= 0 {
			setFields = append(setFields, bson.E{Key: "risk_settings.budget", Value: *rs.Budget})
		}
		if rs.MaxLossPercentage != nil {
			pct := *rs.MaxLossPercentage
			if pct < 0 {
				pct = 0
			}
			if pct > 100 {
				pct = 100
			}
			setFields = append(setFields, bson.E{Key: "risk_settings.max_loss_percentage", Value: pct})
		}
		if rs.RiskTolerance != nil {
			tol := *rs.RiskTolerance
			if tol < 1 {
				tol = 1
			}
			if tol > 10 {
				tol = 10
			}
			setFields = append(setFields, bson.E{Key: "risk_settings.risk_tolerance", Value: tol})
		}
	}

	update := bson.D{{Key: "$set", Value: setFields}}
	_, err = accountsCollection.UpdateOne(ctx, bson.M{"email": email}, update)
	if err != nil {
		logger.Error("failed to update account settings", "err", err)
		util.JSONResponse(res, http.StatusInternalServerError, map[string]any{
			"error": "failed to update settings",
		})
		return
	}

	util.JSONResponse(res, http.StatusOK, map[string]any{
		"success": true,
	})
}
