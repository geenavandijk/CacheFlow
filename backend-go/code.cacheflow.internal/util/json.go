package util

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
)

// JSONResponse writes a JSON response to the client
func JSONResponse(w http.ResponseWriter, status int, payload interface{}) *http.Response {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)

	return &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewReader(response)),
	}
}

// JSONMiddleware is a middleware that ensures the response is JSON
func JSONMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

// ParseJSONBodyMultipart parses the multipart form request body and returns a map
func ParseJSONBodyMultipart(r *http.Request) (map[string]interface{}, error) {
	// Set the logger
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:   true, // Report the file name and line number
		ReportTimestamp: true, // Report the timestamp
		TimeFormat:     "2006-01-02 15:04:05", // Set the time format
		Prefix:         "ACCOUNTS (PJB)", // Set the prefix
	})

	// Ensure the request is parsed as multipart/form-data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		logger.Error("Failed to parse multipart form", err)
		return nil, err
	}

	// Extract the form fields
	result := make(map[string]interface{})
	for key, values := range r.MultipartForm.Value {
		if len(values) > 0 {
			result[key] = values[0] // Assuming single value for each key
		}
	}

	return result, nil
}

func JSONGzipResponse(w http.ResponseWriter, status int, payload interface{}) *http.Response {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzip.NewWriter(w)
	gz.Write(response)
	gz.Close()

	return &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewReader(response)),
	}
}