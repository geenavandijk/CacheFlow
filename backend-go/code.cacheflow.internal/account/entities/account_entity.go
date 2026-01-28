package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountEntity struct {
	ID					primitive.ObjectID 		`json:"_id,omitempty" bson:"_id,omitempty"`
	UserUUID			*string 				`json:"user_uuid,omitempty" bson:"user_uuid,omitempty"`
	IsVerified			*bool 					`json:"is_verified,omitempty" bson:"is_verified,omitempty"`
	IsComplete  		*bool 					`json:"is_complete,omitempty" bson:"is_complete,omitempty"`

	Password			*Password 				`json:"password,omitempty" bson:"password,omitempty"`

	AnnouncementVersion *int64 					`json:"announcement_version" bson:"announcement_version"`

	TwoFAEnabled		*bool 					`json:"two_fa_enabled,omitempty" bson:"two_fa_enabled,omitempty"`
	Sessions			[]*Session 				`json:"sessions,omitempty" bson:"sessions,omitempty"`
	LastLogin           *time.Time 				`json:"last_login,omitempty" bson:"last_login,omitempty"`
	LastLogout          *time.Time 				`json:"last_logout,omitempty" bson:"last_logout,omitempty"`
	FirstName   		*string 				`json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName   			*string 				`json:"last_name,omitempty" bson:"last_name,omitempty"`
	Email				*string 				`json:"email,omitempty" bson:"email,omitempty"`
	CreatedAt   		*time.Time 				`json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   		*time.Time 				`json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Password struct {
	Hash 				*string 				`json:"hash,omitempty" bson:"hash,omitempty"`
	EncryptedVersion	*int64 					`json:"encrypted_version,omitempty" bson:"encrypted_version,omitempty"`
}

type Session struct {
	DeviceUUID			*string 				`json:"device_uuid,omitempty" bson:"device_uuid,omitempty"`
	Token				*string 				`json:"token,omitempty" bson:"token,omitempty"`
	RefreshToken		*string 				`json:"refresh_token,omitempty" bson:"refresh_token,omitempty"`
	RefreshIssuedAt		*time.Time 				`json:"refresh_issued_at,omitempty" bson:"refresh_issued_at,omitempty"`
	TokenID				*string 				`json:"token_id,omitempty" bson:"token_id,omitempty"`
	IssuedAt			*time.Time 				`json:"issued_at,omitempty" bson:"issued_at,omitempty"`
	LastLoginAt			*time.Time 				`json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	UserAgent			*string 				`json:"user_agent,omitempty" bson:"user_agent,omitempty"`
	IPAddress			*string 				`json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	Web 				*bool 					`json:"web,omitempty" bson:"web,omitempty"`
	EncryptedVersion    *int64 					`json:"encrypted_version,omitempty" bson:"encrypted_version,omitempty"`
	EncryptedVersions 	*EncryptedVersions 		`json:"encrypted_versions" bson:"encrypted_versions"`
}

// MARK: EncryptedVersion struct
type EncryptedVersions struct {
	SymmetricVersion  *int64 `json:"symmetric_version" bson:"symmetric_version"`
	AsymmetricVersion *int64 `json:"asymmetric_version" bson:"asymmetric_version"`
}