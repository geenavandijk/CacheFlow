package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return BadRequest("Request body is required", nil)
		}
		return BadRequest("Invalid JSON body", nil).WithErr(err)
	}

	// Ensure no extra garbage after the first JSON value
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return BadRequest("Invalid JSON body", nil)
	}

	return nil
}
