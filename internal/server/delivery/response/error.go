package response

import "net/http"

// Error represents an API error response with multiple possible error messages.
type Error struct {
	// Messages contains one or more error descriptions for the client.
	Messages []string `json:"messages"`
}

// DefaultBadRequestError provides a standard 400 Bad Request error response.
var DefaultBadRequestError = Error{Messages: []string{http.StatusText(http.StatusBadRequest)}}

// DefaultInternalServerError provides a standard 500 Internal Server Error response.
var DefaultInternalServerError = Error{Messages: []string{http.StatusText(http.StatusInternalServerError)}}
