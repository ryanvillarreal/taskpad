package middleware

import "net/http"

// CORS returns middleware that sets Access-Control headers for the given origins.
// If origins is empty, CORS headers are not added. Use "*" for open access (local dev only).
func CORS(origins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if origins == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
