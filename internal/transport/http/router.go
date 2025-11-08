package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"usrsvc/internal/middleware"
)

func NewRouter(h *Handler, allowOrigins []string) http.Handler {
	r := mux.NewRouter()
	r.Use(middleware.CORS(allowOrigins))

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(200) })

	r.HandleFunc("/users", h.ListUsers).Methods(http.MethodGet)
	r.HandleFunc("/users/{id}", h.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/users/{id}", h.UpdateUser).Methods(http.MethodPut)
	r.HandleFunc("/users/{id}", h.DeleteUser).Methods(http.MethodDelete)

	r.HandleFunc("/nationalities", h.ListNationality).Methods(http.MethodGet)
	return r
}
