package routes

import (
	"context"
	"net/http"
	"net/mail"
	"strings"
	"time"

	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/util/httpx"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type CreateAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest

	// Decode JSON
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	// Normalize
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	// Validate
	fields := map[string]string{}

	if req.Email == "" {
		fields["email"] = "required"
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		fields["email"] = "invalid"
	}

	if req.Password == "" {
		fields["password"] = "required"
	} else if len(req.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}

	if len(fields) > 0 {
		httpx.WriteError(w, r, httpx.BadRequest("Validation failed", fields))
		return
	}

	// --- DB create user ---
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users := datastores.DB.Collection("users")

	// Pre-check to give clean 409 even if you don't have a unique index yet
	var existing bson.M
	err := users.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		httpx.WriteError(w, r, httpx.Conflict("Email already in use", map[string]string{
			"email": "already_exists",
		}))
		return
	}
	if err != nil && err != mongo.ErrNoDocuments {
		httpx.WriteError(w, r, httpx.Internal("Internal server error").WithErr(err))
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.WriteError(w, r, httpx.Internal("Internal server error").WithErr(err))
		return
	}

	now := time.Now().UTC()
	doc := bson.M{
		"email":        req.Email,
		"passwordHash": string(hash),
		"createdAt":    now,
		"updatedAt":    now,
		"status":       "active",
	}

	res, err := users.InsertOne(ctx, doc)
	if err != nil {
		// If you have a unique index on email, this catches races too
		if mongo.IsDuplicateKeyError(err) {
			httpx.WriteError(w, r, httpx.Conflict("Email already in use", map[string]string{
				"email": "already_exists",
			}))
			return
		}
		httpx.WriteError(w, r, httpx.Internal("Internal server error").WithErr(err))
		return
	}

	// Return created user id
	idHex := ""
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		idHex = oid.Hex()
	}

	httpx.OK(w, http.StatusCreated, map[string]any{
		"id":    idHex,
		"email": req.Email,
	})
}
