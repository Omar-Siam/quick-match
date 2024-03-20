package authentication

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"net/http"
	"strings"
)

// JWTMiddleware validates the JWT token and extracts the UserID, attaching it to the request context.
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractToken(r)
		if tokenString == "" {
			http.Error(w, "Authorization header is missing or invalid", http.StatusUnauthorized)
			return
		}

		claims := &CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return []byte("quick_match"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, fmt.Sprintf("Invalid or expired token: %v", err), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "UserID", claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts the JWT token from the Authorization header.
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if strings.HasPrefix(bearerToken, "Bearer ") {
		return strings.TrimPrefix(bearerToken, "Bearer ")
	}
	return ""
}
