package routes

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	accountEntities "code.cacheflow.internal/account/entities"
	accountEmail "code.cacheflow.internal/account/handler"
	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/password"
	"code.cacheflow.internal/util/ptr"

	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateAccountBody struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
}

func getLogger(prefix string) *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Prefix:          prefix,
	})
}

func CreateAccount(res http.ResponseWriter, req *http.Request) {
	logger := getLogger("ACCOUNT (CA)")

	var body CreateAccountBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Error("failed to decode request body", "err", err)
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "invalid request body",
		})
		return
	}

	// Basic validation
	if body.Email == nil || strings.TrimSpace(*body.Email) == "" {
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "email is required",
		})
		return
	}
	if body.Password == nil || strings.TrimSpace(*body.Password) == "" {
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "password is required",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(*body.Email))
	ctx := req.Context()

	db := datastores.GetMongoDatabase(ctx)
	accountsCollection := db.Collection(datastores.Accounts)

	// Ensure account doesn't already exist
	var existing accountEntities.AccountEntity
	err := accountsCollection.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&existing)
	switch err {
	case nil:
		util.JSONResponse(res, http.StatusConflict, map[string]any{
			"error": "account already exists",
		})
		return
	case mongo.ErrNoDocuments:
		// ok, proceed to create
	default:
		logger.Error("failed to check existing account", "err", err)
		util.JSONResponse(res, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}

	// Hash password (includes validation)
	hashedPassword, version, err := password.HashPassword(body.Password)
	if err != nil {
		logger.Error("failed to hash password", "err", err)
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	now := time.Now()
	accountID := primitive.NewObjectID()

	account := &accountEntities.AccountEntity{
		ID:         accountID,
		AccountID:  ptr.String(accountID.Hex()),
		IsVerified: ptr.Bool(false),
		IsComplete: ptr.Bool(false),
		Password: &accountEntities.Password{
			Hash:             hashedPassword,
			EncryptedVersion: version,
		},
		AnnouncementVersion: ptr.Int64(0),
		TwoFAEnabled:        ptr.Bool(false),
		Sessions:            []*accountEntities.Session{},
		FirstName:           body.FirstName,
		LastName:            body.LastName,
		Email:               &email,
		CreatedAt:           ptr.Time(now),
		UpdatedAt:           ptr.Time(now),
	}

	if _, err := accountsCollection.InsertOne(ctx, account); err != nil {
		logger.Error("failed to insert account", "err", err)
		util.JSONResponse(res, http.StatusInternalServerError, map[string]any{
			"error": "failed to create account",
		})
		return
	}

	// Create verification record and send email asynchronously
	go func(ctx context.Context, email string, firstName *string) {
		vLogger := getLogger("ACCOUNT (CV)")
		verificationSecret, err := generateVerificationSecret()
		if err != nil {
			vLogger.Error("failed to generate verification secret", "err", err)
			return
		}

		db := datastores.GetMongoDatabase(ctx)
		verificationCollection := db.Collection(datastores.AccountCreationVerification)

		now := time.Now()
		verification := &accountEntities.VerificationEntity{
			ID:         primitive.NewObjectID(),
			UUID:       ptr.String(verificationSecret),
			IsVerified: ptr.Bool(false),
			Resends:    ptr.Int32(0),
			IsComplete: ptr.Bool(false),
			DeviceID:   nil,
			Code:       ptr.Int64(0),
			Info:       ptr.String(email), // store email for lookup
			CreatedAt:  ptr.Time(now),
			Attempts:   ptr.Int32(0),
		}

		if _, err := verificationCollection.InsertOne(ctx, verification); err != nil {
			vLogger.Error("failed to insert verification record", "err", err)
			return
		}

		verificationLink := fmt.Sprintf("http://localhost:5173/verify?secret=%s", verificationSecret)

		emailData := map[string]string{
			"verification_link": verificationLink,
			"email":             email,
		}
		if firstName != nil {
			emailData["first_name"] = *firstName
		}

		if err := accountEmail.SendEmail(accountEmail.EmailRequestBody{
			Email:    email,
			Subject:  "",
			Template: "verify-create-account",
			Data:     emailData,
		}); err != nil {
			vLogger.Error("failed to send verification email", "err", err)
		}
	}(context.Background(), email, body.FirstName)

	util.JSONResponse(res, http.StatusCreated, map[string]any{
		"success": true,
		"email":   email,
	})
}

func VerifyAccount(res http.ResponseWriter, req *http.Request) {
	logger := getLogger("ACCOUNT (VA)")
	ctx := req.Context()

	secret := req.URL.Query().Get("secret")
	if strings.TrimSpace(secret) == "" {
		util.JSONResponse(res, http.StatusBadRequest, map[string]any{
			"error": "secret is required",
		})
		return
	}

	db := datastores.GetMongoDatabase(ctx)
	verificationCollection := db.Collection(datastores.AccountCreationVerification)

	var verification accountEntities.VerificationEntity
	err := verificationCollection.FindOne(ctx, bson.D{{Key: "uuid", Value: secret}}).Decode(&verification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			util.JSONResponse(res, http.StatusNotFound, map[string]any{
				"error": "invalid or expired verification link",
			})
			return
		}
		logger.Error("failed to find verification record", "err", err)
		util.JSONResponse(res, http.StatusInternalServerError, map[string]any{
			"error": "internal server error",
		})
		return
	}

	if verification.Info == nil || strings.TrimSpace(*verification.Info) == "" {
		util.JSONResponse(res, http.StatusInternalServerError, map[string]any{
			"error": "verification record is missing account info",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(*verification.Info))
	accountsCollection := db.Collection(datastores.Accounts)

	// Mark account as verified and complete
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "is_verified", Value: true},
			{Key: "is_complete", Value: true},
			{Key: "updated_at", Value: time.Now()},
		},
	}}

	if _, err := accountsCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: email}}, update); err != nil {
		logger.Error("failed to update account verification status", "err", err)
		util.JSONResponse(res, http.StatusInternalServerError, map[string]any{
			"error": "failed to verify account",
		})
		return
	}

	// Delete verification record now that it's been used
	if _, err := verificationCollection.DeleteOne(ctx, bson.D{{Key: "_id", Value: verification.ID}}); err != nil {
		logger.Error("failed to delete verification record", "err", err)
	}

	util.JSONResponse(res, http.StatusOK, map[string]any{
		"success": true,
		"email":   email,
	})
}

func generateVerificationSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buf), nil
}
