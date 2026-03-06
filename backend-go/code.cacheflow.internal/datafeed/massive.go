package datafeed

import (
	"code.cacheflow.internal/util/secrets"
	massive "github.com/massive-com/client-go/v2/rest"
)

func GetMassiveClient() *massive.Client {
	c := massive.New(secrets.MassiveMainApiKeyValue)
	
	return c
}