package http

import "net/http"

const (
	StatusOK                  = http.StatusOK // 200
	StatusCreated             = http.StatusCreated
	StatusBadRequest          = http.StatusBadRequest          // 400
	StatusUnprocessableEntity = http.StatusUnprocessableEntity // 422
	StatusNotFound            = http.StatusNotFound            // 404
	StatusInternalServerError = http.StatusInternalServerError // 500
	StatusConflict            = http.StatusConflict
)
