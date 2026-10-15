package utils

import (
	"net/http/httptest"
	"testing"
)

func TestTokens(t *testing.T) {
	secret := "testsecret12345"
	userID := 42
	role := "seller"

	tokenStr, err := GenerateAccessToken(userID, role, secret)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := ParseAccessToken(tokenStr, secret)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}

	refreshTokenStr, err := GenerateRefreshToken(userID, secret)
	if err != nil {
		t.Fatalf("unexpected error generating refresh token: %v", err)
	}

	refreshClaims, err := ParseRefreshToken(refreshTokenStr, secret)
	if err != nil {
		t.Fatalf("unexpected error validating refresh token: %v", err)
	}

	if refreshClaims.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, refreshClaims.UserID)
	}
}

func TestSendJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	payload := map[string]string{"foo": "bar"}
	SendJSON(recorder, 201, payload)

	if recorder.Code != 201 {
		t.Errorf("expected status code 201, got %d", recorder.Code)
	}

	expectedBody := `{"data":{"foo":"bar"}}` + "\n"
	if recorder.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, recorder.Body.String())
	}
}

func TestSendError(t *testing.T) {
	recorder := httptest.NewRecorder()
	SendError(recorder, 400, "something went wrong")

	if recorder.Code != 400 {
		t.Errorf("expected 400, got %d", recorder.Code)
	}

	expectedBody := `{"error":"something went wrong"}` + "\n"
	if recorder.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, recorder.Body.String())
	}
}

func TestInvalidTokens(t *testing.T) {
	secret := "testsecret12345"
	wrongSecret := "wrongsecret12345"

	tokenStr, _ := GenerateAccessToken(42, "seller", secret)
	_, err := ParseAccessToken(tokenStr, wrongSecret)
	if err == nil {
		t.Error("expected error for token with wrong secret, got nil")
	}

	_, err = ParseAccessToken("invalidtoken", secret)
	if err == nil {
		t.Error("expected error for invalid token format, got nil")
	}

	refreshTokenStr, _ := GenerateRefreshToken(42, secret)
	_, err = ParseRefreshToken(refreshTokenStr, wrongSecret)
	if err == nil {
		t.Error("expected error for refresh token with wrong secret, got nil")
	}

	_, err = ParseRefreshToken("invalidtoken", secret)
	if err == nil {
		t.Error("expected error for invalid refresh token format, got nil")
	}
}
