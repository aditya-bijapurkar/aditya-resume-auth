package auth

import (
	"context"
	"net/http"

	"github.com/aditya-bijapurkar/aditya-resume-auth/utils"
)

type contextKey string

const UserIDKey contextKey = "auth_user_id"
const UserEmailKey contextKey = "auth_email"
const AuthTokenCookieName = "go_auth_token"
const InternalTokenHeaderName = "Internal-Token"

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie(AuthTokenCookieName)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := cookie.Value

		claims, err := utils.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func VerifyTokenSignature(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(InternalTokenHeaderName)
		if token == "" {
			http.Error(w, "Missing "+InternalTokenHeaderName+" header", http.StatusUnauthorized)
			return
		}

		valid, err := utils.ValidateSignature(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func AddAuthTokenContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AuthTokenCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		token := cookie.Value
		claims, err := utils.ValidateToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
