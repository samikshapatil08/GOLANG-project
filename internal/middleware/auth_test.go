package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"switchboard/internal/utils"
)

func TestRequireAuth(t *testing.T) {
	secret := "mysecretkey"
	userID := 123
	role := "buyer"

	validToken, err := utils.GenerateAccessToken(userID, role, secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	testHandler := http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		uid, ok := GetUserID(req.Context())
		if !ok || uid != userID {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token value",
			authHeader:     "Bearer invalidtokenhere",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			recorder := httptest.NewRecorder()

			middlewareFn := RequireAuth(secret)
			middlewareFn(testHandler).ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestGetRole(t *testing.T) {
	ctx := context.Background()
	if _, ok := GetRole(ctx); ok {
		t.Error("expected ok to be false for empty context")
	}

	ctx = context.WithValue(ctx, RoleKey, "admin")
	role, ok := GetRole(ctx)
	if !ok || role != "admin" {
		t.Errorf("expected role 'admin', got %q", role)
	}
}

func TestLogger(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	recorder := httptest.NewRecorder()

	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTeapot)
	})

	Logger(next).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTeapot {
		t.Errorf("expected 418 status, got %d", recorder.Code)
	}
}
