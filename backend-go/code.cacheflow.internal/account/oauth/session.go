package oauth

import (
	"context"
	"crypto/rand"
	"crypto/x509"

	"encoding/pem"
	"errors"
	"fmt"

	"net"
	"net/http"
	"os"
	"strings"
	"time"

	accountModel "code.cacheflow.internal/account/entities"
	"code.cacheflow.internal/apicontext"
	datastores "code.cacheflow.internal/datastores/mongo"

	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/ptr"
	"code.cacheflow.internal/util/secrets"
	"github.com/charmbracelet/log"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getIP(r *http.Request) string {
	// Check if the request was forwarded by a proxy
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For might contain a list of IPs, take the first one
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])
		return clientIP
	}

	// If no X-Forwarded-For header, get the IP from RemoteAddr
	ip := r.RemoteAddr

	// In some cases, RemoteAddr might include a port, so you may want to remove it
	// Split the IP and port
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip // return the whole RemoteAddr if there's an error
	}

	return host
}

// Token, TokenID, RefreshToken, TokenVersion, RefreshTokenVersion, Error
func GenerateOAuth2Token(req *http.Request, res http.ResponseWriter, ctx context.Context) (*string, *string, *string, *int32, *int32, error) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,                  // Report the file name and line number
		ReportTimestamp: true,                  // Report the timestamp
		TimeFormat:      "2006-01-02 15:04:05", // Set the time format
		Prefix:          "AUTH SESSION (IT)",   // Set the prefix
	})

	deviceID := req.Header.Get("x-cf-device-id")
	userAgent := req.Header.Get("User-Agent")
	email := req.Header.Get("x-cf-uid")
	email = strings.ToLower(email)

	if deviceID == "" || userAgent == "" || email == "" {
		return nil, nil, nil, nil, nil, errors.New("missing required headers")
	}

	var account accountModel.AccountEntity
	accountsCollection := datastores.GetMongoDatabase(ctx).Collection("accounts")

	filter := bson.D{{Key: "email", Value: email}}
	err := accountsCollection.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments && !util.ContainsString(req.URL.Path, "/exists") {
			return nil, nil, nil, nil, nil, errors.New("account not found")
		} else if err != mongo.ErrNoDocuments && !util.ContainsString(req.URL.Path, "/exists") {
			logger.Error("Internal server error 1 : " + err.Error())
			return nil, nil, nil, nil, nil, errors.New("account not found")
		}
	}

	privateKeyLatestVersion := secrets.PrivateKeyValueLatestVersion

	// Get the latest private key and version
	privateKey, exists := secrets.GetPrivateKeyWithVersion(secrets.PrivateKeyValueLatestVersion)
	if !exists {
		return nil, nil, nil, nil, nil, errors.New("private key not found")
	}

	// Generate a new token
	var audience []string
	audience = []string{
		"/v1",
	}

	accessToken, tokenID, err := GenerateToken(deviceID, &account, privateKey, req.Header.Get("x-cf-uid"), audience)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// Generate refresh token
	refreshToken := make([]byte, 32)
	_, err = rand.Read(refreshToken)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	refreshTokenString := fmt.Sprintf("%x", refreshToken)

	if account.AccountID != nil {

		var found bool
		sessions := account.Sessions

		// Loop by index so we update the actual slice element, not a copy
		for i := range sessions {
			session := sessions[i]
			if session != nil && *session.DeviceID == deviceID {
				found = true
				// Update the session in place so MongoDB gets the new token
				session.LastLoginAt = ptr.Time(time.Now())
				session.IssuedAt = ptr.Time(time.Now())
				session.RefreshToken = &refreshTokenString
				session.RefreshIssuedAt = ptr.Time(time.Now())
				session.Token = &accessToken
				session.TokenID = &tokenID
				session.EncryptionVersions = &accountModel.EncryptedVersions{
					AsymmetricVersion: ptr.Int64(int64(privateKeyLatestVersion)),
					SymmetricVersion:  ptr.Int64(int64(privateKeyLatestVersion)),
				}
				account.Sessions = sessions
				break
			}
		}

		if !found {

			// Create a new session
			session := &accountModel.Session{
				DeviceID:        &deviceID,
				Token:           &accessToken,
				TokenID:         &tokenID,
				RefreshToken:    &refreshTokenString,
				RefreshIssuedAt: ptr.Time(time.Now()),
				IssuedAt:        ptr.Time(time.Now()),
				LastLoginAt:     ptr.Time(time.Now()),
				IPAddress:       ptr.String(getIP(req)),
				Active:          ptr.Bool(true),
				EncryptionVersions: &accountModel.EncryptedVersions{
					AsymmetricVersion: ptr.Int64(int64(privateKeyLatestVersion)),
					SymmetricVersion:  ptr.Int64(int64(privateKeyLatestVersion)),
				},
			}

			// Append the session to the account
			account.Sessions = append(account.Sessions, session)
		}

		account.LastLogin = ptr.Time(time.Now())

		// Update the account
		_, err = accountsCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: account}})
		if err != nil {
			logger.Error("Internal server error 2 : " + err.Error())
			return nil, nil, nil, nil, nil, errors.New("internal server error")
		}
	}

	return &accessToken, &tokenID, &refreshTokenString, ptr.Int32(int32(privateKeyLatestVersion)), ptr.Int32(int32(privateKeyLatestVersion)), nil
}

// Access token, Refresh token, Token ID, Token Version, Refresh Token Version, Error
func RefreshOAuth2Token(scope string, req *http.Request, res http.ResponseWriter, refreshToken *string) (*string, *string, *string, *int32, *int32, error) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,                  // Report the file name and line number
		ReportTimestamp: true,                  // Report the timestamp
		TimeFormat:      "2006-01-02 15:04:05", // Set the time format
		Prefix:          "AUTH SESSION (IT)",   // Set the prefix
	})

	deviceID := req.Header.Get("x-cf-device-id")
	userAgent := req.Header.Get("User-Agent")
	email := req.Header.Get("x-cf-uid")
	email = strings.ToLower(email)

	if deviceID == "" || userAgent == "" || email == "" {
		return nil, nil, nil, nil, nil, errors.New("missing required headers")
	}

	var account accountModel.AccountEntity
	accountsCollection := datastores.GetMongoDatabase(context.Background()).Collection(datastores.Accounts)

	filter := bson.D{{Key: "email", Value: email}}
	err := accountsCollection.FindOne(context.Background(), filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments && !util.ContainsString(req.URL.Path, "/exists") {
			return nil, nil, nil, nil, nil, errors.New("account not found")
		} else if err != mongo.ErrNoDocuments && !util.ContainsString(req.URL.Path, "/exists") {
			logger.Error("Internal server error 1 : " + err.Error())
			return nil, nil, nil, nil, nil, errors.New("account not found")
		}
	}

	privateKeyLatestVersion := secrets.PrivateKeyValueLatestVersion

	// Find the verified device
	sessions := account.Sessions
	var found bool
	var foundVerifiedDevice accountModel.Session

	for _, session := range sessions {
		if *session.DeviceID == deviceID {
			found = true
			foundVerifiedDevice = *session
		}
	}

	if !found {
		logger.Error("No session found for the following account: ", email)
		return nil, nil, nil, nil, nil, errors.New("no session found for the following account: " + email)
	}

	// Check if the refresh token is valid
	if *foundVerifiedDevice.RefreshToken != *refreshToken {
		logger.Info("Found refresh token: ", *foundVerifiedDevice.RefreshToken)
		logger.Info("Current refresh token: ", *refreshToken)
		logger.Error("Invalid refresh token for the following account: ", email)
		return nil, nil, nil, nil, nil, errors.New("invalid refresh token for the following account: " + email)
	}

	accessToken, tokenID, newRefreshTokenString, _, _, err := GenerateOAuth2Token(req, res, context.Background())

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	fmt.Println("New refresh token: ", newRefreshTokenString)
	fmt.Println("old refresh token: ", *refreshToken)

	return accessToken, newRefreshTokenString, tokenID, ptr.Int32(int32(privateKeyLatestVersion)), ptr.Int32(int32(privateKeyLatestVersion)), nil
}

func RevokeOAuth2Token(req *http.Request, res http.ResponseWriter, ctx context.Context) error {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,                  // Report the file name and line number
		ReportTimestamp: true,                  // Report the timestamp
		TimeFormat:      "2006-01-02 15:04:05", // Set the time format
		Prefix:          "AUTH SESSION (RT)",   // Set the prefix
	})

	deviceID := req.Header.Get("x-cf-device-id")
	email := req.Header.Get("x-cf-uid")
	email = strings.ToLower(email)

	if deviceID == "" || email == "" {
		return errors.New("missing required headers")
	}

	var account accountModel.AccountEntity
	accountsCollection := datastores.GetMongoDatabase(context.Background()).Collection(datastores.Accounts)

	filter := bson.D{{Key: "email", Value: email}}
	err := accountsCollection.FindOne(context.Background(), filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments && !util.ContainsString(req.URL.Path, "/exists") {
			return errors.New("account not found")
		} else if err != mongo.ErrNoDocuments && !util.ContainsString(req.URL.Path, "/exists") {
			logger.Error("Internal server error 1 : " + err.Error())
			return errors.New("account not found")
		}
	}

	// Find the verified device
	sessions := account.Sessions
	var found bool
	var foundVerifiedDevice accountModel.Session

	for _, session := range sessions {
		if *session.DeviceID == deviceID {
			found = true
			foundVerifiedDevice = *session
		}
	}

	if !found {
		logger.Error("No session found for the following account: ", email)
		return errors.New("no session found for the following account: " + email)
	}

	// Update the session
	foundVerifiedDevice.Token = nil
	foundVerifiedDevice.RefreshToken = nil
	foundVerifiedDevice.RefreshIssuedAt = nil
	foundVerifiedDevice.TokenID = nil
	foundVerifiedDevice.LastLoginAt = ptr.Time(time.Now())
	foundVerifiedDevice.EncryptionVersions = nil
	foundVerifiedDevice.Active = ptr.Bool(false)

	// Set the session
	for i, session := range sessions {
		if *session.DeviceID == deviceID {
			sessions[i] = &foundVerifiedDevice
		}
	}

	// Update the account
	account.Sessions = sessions

	// Update the account in MongoDB
	_, err = accountsCollection.UpdateOne(context.Background(), filter, bson.D{{Key: "$set", Value: account}})
	if err != nil {
		logger.Error("Error updating account: ", err)
		return err
	}

	return nil
}

// MARK: VerifyToken
func VerifyOAuthToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		uid := req.Header.Get("x-cf-uid")
		uid = strings.ToLower(uid)

		fmt.Println("Looking for email (uid): ", uid)

		cDeviceID := req.Header.Get("x-cf-device-id")

		if uid == "" || cDeviceID == "" {
			util.JSONResponse(res, http.StatusUnauthorized, map[string]interface{}{
				"error": "x-cf-uid or x-cf-device-id is empty",
			})
			return
		}

		token := req.Header.Get("x-cf-bearer")
		if token == "" {
			util.JSONResponse(res, http.StatusUnauthorized, map[string]interface{}{
				"error": "Token is empty",
			})
			return
		}

		newCtx, err := VerifyOAuthTokenInternal(req, res, req.Context())
		if err != nil {
			util.JSONResponse(res, http.StatusUnauthorized, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		// Update request with new context that contains the account object
		req = req.WithContext(newCtx)

		next.ServeHTTP(res, req) // Call the next handler with updated request
	})
}

func VerifyOAuthTokenInternal(req *http.Request, res http.ResponseWriter, ctx context.Context) (context.Context, error) {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Prefix:          "AUTH SESSION (VTI)",
	})

	cFrom := req.Header.Get("x-cf-uid")
	cDeviceID := req.Header.Get("x-cf-device-id")

	cFrom = strings.ToLower(cFrom)

	if cFrom == "" {
		// get from query for websocket
		cFrom = req.URL.Query().Get("username")
	}

	if cDeviceID == "" {
		// get from query for websocket
		cDeviceID = req.URL.Query().Get("device_id")
	}

	token := req.Header.Get("x-cf-bearer")
	if token == "" {
		return ctx, errors.New("token is empty")
	}

	var account accountModel.AccountEntity
	accountsCollection := datastores.GetMongoDatabase(ctx).Collection(datastores.Accounts)
	// Use context.Background() for database query to avoid request context timeouts
	dbCtx := context.Background()
	err := accountsCollection.FindOne(dbCtx, bson.D{{Key: "email", Value: cFrom}}, options.FindOne().SetProjection(bson.D{
		{Key: "sessions", Value: 1},
		{Key: "email", Value: 1},
		{Key: "account_id", Value: 1},
	}).SetHint(bson.D{{Key: "email", Value: 1}})).Decode(&account)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx, errors.New("no account found with the following email: " + cFrom)
		}
		logger.Error("Error finding account 1: ", err)
		logger.Error("Error type: ", fmt.Sprintf("%T", err))
		logger.Error("Error details - email: ", cFrom)
		return ctx, errors.New("error finding account 1 with email: " + cFrom + " - " + err.Error())
	}

	// Verify session + token
	var foundVerifiedDevice accountModel.Session
	var found bool
	for _, session := range account.Sessions {
		if *session.DeviceID == cDeviceID {
			found = true
			foundVerifiedDevice = *session
			break
		}
	}
	if !found {
		logger.Error("No session found for deviceID")
		return ctx, errors.New("no session found")
	}
	if *foundVerifiedDevice.Token != token {
		logger.Error("Invalid token")
		return ctx, errors.New("invalid token")
	}

	// Public key and JWT verification
	publicKey, retrieved := secrets.GetPublicKeyWithVersion(int32(*foundVerifiedDevice.EncryptionVersions.SymmetricVersion))
	if !retrieved {
		logger.Error("Error retrieving public key")
		return ctx, errors.New("public key version does not exist")
	}
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		logger.Error("PEM decode failed")
		return ctx, errors.New("no PEM data found")
	}
	rsaPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logger.Error("Public key parse failed")
		return ctx, errors.New("error parsing public key")
	}

	tokenVerify, err := jwt.Parse(*foundVerifiedDevice.Token, func(token *jwt.Token) (interface{}, error) {
		return rsaPublicKey, nil
	})
	if err != nil {
		logger.Error("JWT verify failed", err)
		return ctx, errors.New("invalid or expired token")
	}
	claims := tokenVerify.Claims.(jwt.MapClaims)
	if time.Now().Unix() > int64(claims["exp"].(float64)) {
		logger.Error("Token expired")
		return ctx, errors.New("session expired")
	}

	// Authorization check
	audience := []string{
		"/v1",
	}
	isAuthorized := checkAuthorization(req.URL.String(), audience)

	if !isAuthorized {
		logger.Error("Unauthorized endpoint")
		return ctx, errors.New("unauthorized")
	}

	go updateLastLoginAt(account, account.Sessions, foundVerifiedDevice)
	ctx = context.WithValue(ctx, apicontext.AccountContextKey, account)

	req = req.WithContext(ctx)
	_ = req
	return ctx, nil
}

func updateLastLoginAt(account accountModel.AccountEntity, sessions []*accountModel.Session, foundVerifiedDevice accountModel.Session) {
	// Perform the session update asynchronously after returning the response
	foundVerifiedDevice.LastLoginAt = ptr.Time(time.Now())
	_ = foundVerifiedDevice.LastLoginAt
	account.Sessions = sessions
	_ = account.Sessions

	// Update the account in MongoDB
	accountsCollection := datastores.GetMongoDatabase(context.Background()).Collection("accounts")
	filter := bson.D{{Key: "email", Value: account.Email}}
	_, err := accountsCollection.UpdateOne(context.Background(), filter, bson.D{{Key: "$set", Value: bson.D{
		{Key: "sessions", Value: sessions},
	}}})
	if err != nil {
		fmt.Println("Error updating account: ", err)
	}
}

// MARK: GenerateToken
func GenerateToken(deviceID string, account *accountModel.AccountEntity, privateKey string, username string, audience []string) (string, string, error) {

	// Generate token ID (8 digits)
	tokenID := make([]byte, 8)
	_, err := rand.Read(tokenID)
	if err != nil {
		return "", "", err
	}

	// Convert token ID to string
	tokenIDString := fmt.Sprintf("%x", tokenID)

	// Parse the private key
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return "", "", errors.New("error decoding private key")
	}

	// Parse the private key
	privateKeyParsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", "", err
	}

	// Generate a new token
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["aud"] = audience
	claims["sub"] = username
	claims["account_id"] = *account.AccountID
	claims["device_id"] = deviceID

	// Set expiration to 24 hours from now
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix() // Expiration time as a Unix timestamp
	// claims["exp"] = time.Now().Add(1 * time.Minute).Unix() // Expiration time as a Unix timestamp
	claims["iat"] = time.Now().Unix() // Issued at timestamp
	claims["jti"] = tokenIDString     // Unique token ID

	// Sign the token
	tokenString, err := token.SignedString(privateKeyParsed)
	if err != nil {
		return "", "", err
	}

	return tokenString, tokenIDString, nil
}

func checkAuthorization(url string, audience []string) bool {
	for _, route := range audience {
		if strings.Contains(url, route) {
			return true
		}
	}
	return false
}
