package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MARK: Verification struct
// This handles verifying the account email on sign up
type VerificationEntity struct {
	ID		  		primitive.ObjectID 		`json:"_id" bson:"_id"`
	UUID	  		*string 				`json:"uuid" bson:"uuid"`
	IsVerified 		*bool 					`json:"is_verified" bson:"is_verified"`
	Resends 		*int32 					`json:"resends" bson:"resends"`
	IsComplete 		*bool 					`json:"is_complete" bson:"is_complete"`
	DeviceID 		*string 				`json:"device_id" bson:"device_id"`
	Code 			*int64 					`json:"code" bson:"code"`
	Info 			*string 				`json:"info" bson:"info"`
	CreatedAt 		*time.Time 				`json:"created_at" bson:"created_at"`
	Attempts 		*int32 					`json:"attempts" bson:"attempts"`
}