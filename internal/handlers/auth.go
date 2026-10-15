package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"switchboard/internal/models"
	"switchboard/internal/repository"
	"switchboard/internal/utils"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=buyer seller"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken string      `json:"access_token"`
	User        UserSummary `json:"user"`
}

type UserSummary struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (handler *Handler) Register(writer http.ResponseWriter, req *http.Request) {
	var requestBody RegisterRequest
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := handler.validate.Struct(requestBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := &models.User{
		Name:     requestBody.Name,
		Email:    requestBody.Email,
		Password: string(hashedPassword),
		Role:     requestBody.Role,
	}

	err = handler.Users.CreateUser(req.Context(), user)
	if err != nil {
		utils.SendError(writer, http.StatusConflict, err.Error())
		return
	}

	summary := UserSummary{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}

	utils.SendJSON(writer, http.StatusCreated, summary)
}

func (handler *Handler) Login(writer http.ResponseWriter, req *http.Request) {
	var requestBody LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := handler.validate.Struct(requestBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, err.Error())
		return
	}

	user, err := handler.Users.GetUserByEmail(req.Context(), requestBody.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.SendError(writer, http.StatusUnauthorized, "invalid credentials")
			return
		}
		utils.SendError(writer, http.StatusInternalServerError, "internal server error")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestBody.Password))
	if err != nil {
		utils.SendError(writer, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role, handler.JWTSecret)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, handler.JWTSecret)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	}
	http.SetCookie(writer, cookie)

	summary := UserSummary{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}

	response := AuthResponse{
		AccessToken: accessToken,
		User:        summary,
	}

	utils.SendJSON(writer, http.StatusOK, response)
}

func (handler *Handler) Refresh(writer http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("refresh_token")
	if err != nil {
		utils.SendError(writer, http.StatusUnauthorized, "missing refresh token cookie")
		return
	}

	claims, err := utils.ParseRefreshToken(cookie.Value, handler.JWTSecret)
	if err != nil {
		utils.SendError(writer, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	user, err := handler.Users.GetUserByID(req.Context(), claims.UserID)
	if err != nil {
		utils.SendError(writer, http.StatusUnauthorized, "user not found")
		return
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Role, handler.JWTSecret)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response := RefreshResponse{
		AccessToken: newAccessToken,
	}

	utils.SendJSON(writer, http.StatusOK, response)
}

func (handler *Handler) Logout(writer http.ResponseWriter, req *http.Request) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(writer, cookie)

	utils.SendJSON(writer, http.StatusOK, map[string]string{"message": "logged out successfully"})
}
