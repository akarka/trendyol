package server

import (
	"crypto/subtle"
	"net/http"
)

// basicAuth, webhook endpoint'ini Trendyol kimlik bilgileriyle korur.
func basicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			userOK := subtle.ConstantTimeCompare([]byte(u), []byte(username)) == 1
			passOK := subtle.ConstantTimeCompare([]byte(p), []byte(password)) == 1
			if !ok || !userOK || !passOK {
				w.Header().Set("WWW-Authenticate", `Basic realm="webhook"`)
				http.Error(w, "yetkisiz", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
