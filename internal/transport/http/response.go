package http

import (
	"encoding/json"
	"net/http"
)

type apiError struct {
	Error   bool              `json:"error"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string, fields map[string]string) {
	writeJSON(w, status, apiError{Error: true, Message: msg, Fields: fields})
}
