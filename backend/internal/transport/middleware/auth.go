package middleware

import (
	"net/http"
	"strings"
)

func AuthBearer(expectedToken string, excludedUrls map[string]struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := excludedUrls[r.URL.Path]; ok {
				// Skip auth for excluded URLs
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "

			if !strings.HasPrefix(auth, prefix) {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			got := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
			if got != expectedToken {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
