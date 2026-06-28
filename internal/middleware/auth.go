package middleware

import (
	"context"
	"net/http"
	"strings"

	"switchboard/internal/utils"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				utils.SendError(writer, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.SendError(writer, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenStr := parts[1]
			claims, err := utils.ParseAccessToken(tokenStr, jwtSecret)
			if err != nil {
				utils.SendError(writer, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := req.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(writer, req.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (int, bool) {
	val, ok := ctx.Value(UserIDKey).(int)
	return val, ok
}

func GetRole(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(RoleKey).(string)
	return val, ok
}
