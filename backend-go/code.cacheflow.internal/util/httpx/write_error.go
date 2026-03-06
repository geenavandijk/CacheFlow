package httpx

import (
	"log"
	"net/http"
)

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	rid := RequestID(r)
	status, apiErr := ToAPIError(err, rid)

	// Log server-side errors (500s) with details
	if status >= 500 {
		log.Printf("error request_id=%s err=%v", rid, err)
	}

	WriteJSON(w, status, APIResponse{Ok: false, Error: apiErr})
}
