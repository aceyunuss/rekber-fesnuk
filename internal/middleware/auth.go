package middleware

import (
	"context"
	"net/http"
	"rekber-fesnuk/internal/auth"
	"rekber-fesnuk/internal/httputil"
	"strings"
)

type contextKey string

const SellerIDKey contextKey = "seller_id"

func ReqAuth(jwtSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if !strings.HasPrefix(authHeader, "Bearer ") {
				httputil.WriteError(w, http.StatusUnauthorized, "No token")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := auth.ValidateToken(jwtSecret, tokenString)
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), SellerIDKey, claims.SellerId)
			next(w, r.WithContext(ctx))
		}
	}
}
