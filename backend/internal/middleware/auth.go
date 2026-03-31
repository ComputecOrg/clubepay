package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/clubepay/backend/internal/domain"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

// AuthCookieName é o nome do cookie HttpOnly que armazena o JWT de autenticação.
const AuthCookieName = "assinapix_token"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				writeErr(w, http.StatusUnauthorized, "missing authorization")
				return
			}

			claims, err := domain.ParseJWT(tokenStr, jwtSecret)
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken lê o token JWT do cookie HttpOnly primeiro, com fallback para o header Authorization.
func extractToken(r *http.Request) string {
	if cookie, err := r.Cookie(AuthCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if RoleFromContext(r.Context()) != role {
				writeErr(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(userIDKey).(int64)
	return id
}

func RoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(roleKey).(string)
	return role
}

// UserIDContextKey exports the context key so handler tests can build authenticated contexts directly.
func UserIDContextKey() contextKey { return userIDKey }

// RoleContextKey exports the context key so handler tests can build authenticated contexts directly.
func RoleContextKey() contextKey { return roleKey }

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
