package middleware

import "net/http"

func CORS(allow []string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, o := range allow { allowed[o] = struct{}{} }
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if _, ok := allowed[origin]; ok || len(allowed) == 0 {
				if origin == "" { origin = "*" }
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			}
			if r.Method == "OPTIONS" { w.WriteHeader(204); return }
			next.ServeHTTP(w, r)
		})
	}
}
