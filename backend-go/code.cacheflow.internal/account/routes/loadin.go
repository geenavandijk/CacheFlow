package routes

import (
	"net/http"

	accountEntities "code.cacheflow.internal/account/entities"
	datastores "code.cacheflow.internal/datastores/mongo"
	"code.cacheflow.internal/util"
	"code.cacheflow.internal/util/httpx"
	"go.mongodb.org/mongo-driver/bson"
)

func LoadInAccountData(res http.ResponseWriter, req *http.Request) {

	email := req.Header.Get("x-cf-uid")
	if email == "" {
		httpx.WriteError(res, req, httpx.BadRequest("email is required", map[string]string{
			"email": email,
		}))
		return
	}

	accountCollection := datastores.GetMongoDatabase(req.Context()).Collection(datastores.Accounts)

	var account *accountEntities.AccountEntity
	err := accountCollection.FindOne(req.Context(), bson.M{"email": email}).Decode(&account)
	if err != nil {
		httpx.WriteError(res, req, httpx.BadRequest("account not found", map[string]string{
			"email": email,
		}))
		return
	}

	util.JSONResponse(res, http.StatusOK, map[string]any{
		"email": account.Email,
		"first_name": account.FirstName,
		"last_name": account.LastName,
		"account_id": account.AccountID,
	})
}