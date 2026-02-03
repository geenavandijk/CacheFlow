package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountEntity struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AccountID  *string            `json:"account_id,omitempty" bson:"account_id,omitempty"`
	IsVerified *bool              `json:"is_verified,omitempty" bson:"is_verified,omitempty"`
	IsComplete *bool              `json:"is_complete,omitempty" bson:"is_complete,omitempty"`

	Password *Password `json:"password,omitempty" bson:"password,omitempty"`

	AnnouncementVersion *int64 `json:"announcement_version" bson:"announcement_version"`

	TwoFAEnabled *bool      `json:"two_fa_enabled,omitempty" bson:"two_fa_enabled,omitempty"`
	Sessions     []*Session `json:"sessions,omitempty" bson:"sessions,omitempty"`
	LastLogin    *time.Time `json:"last_login,omitempty" bson:"last_login,omitempty"`
	LastLogout   *time.Time `json:"last_logout,omitempty" bson:"last_logout,omitempty"`
	FirstName    *string    `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName     *string    `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Email        *string    `json:"email,omitempty" bson:"email,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Password struct {
	Hash             *string `json:"hash,omitempty" bson:"hash,omitempty"`
	EncryptedVersion *int64  `json:"encrypted_version,omitempty" bson:"encrypted_version,omitempty"`
}

type Session struct {
	DeviceID           *string            `json:"device_id" bson:"device_id"`                     // device id
	Token              *string            `json:"token" bson:"token"`                             // jwt token
	RefreshToken       *string            `json:"refresh_token" bson:"refresh_token"`             // refresh token
	RefreshIssuedAt    *time.Time         `json:"refresh_issued_at" bson:"refresh_issued_at"`     // refresh token issued at
	TokenID            *string            `json:"token_id" bson:"token_id"`                       // session token id
	IssuedAt           *time.Time         `json:"issued_at" bson:"issued_at"`                     // session token issued at
	LastLoginAt        *time.Time         `json:"last_login_at" bson:"last_login_at"`             // last login at
	Active             *bool              `json:"active" bson:"active"`                           // flag for active session
	IPAddress          *string            `json:"ip_address" bson:"ip_address"`                   // ip address of the session
	EncryptionVersions *EncryptedVersions `json:"encryption_versions" bson:"encryption_versions"` // encryption versions struct
}

// MARK: EncryptedVersion struct
type EncryptedVersions struct {
	SymmetricVersion  *int64 `json:"symmetric_version" bson:"symmetric_version"`
	AsymmetricVersion *int64 `json:"asymmetric_version" bson:"asymmetric_version"`
}
