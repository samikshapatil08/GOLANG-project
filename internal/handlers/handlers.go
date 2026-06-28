package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"switchboard/internal/repository"
	"switchboard/internal/utils"
)

type Handler struct {
	Listings  repository.ListingRepository
	Users     repository.UserRepository
	JWTSecret string
	validate  *validator.Validate
}

func NewHandler(listings repository.ListingRepository, users repository.UserRepository, jwtSecret string) *Handler {
	return &Handler{
		Listings:  listings,
		Users:     users,
		JWTSecret: jwtSecret,
		validate:  validator.New(),
	}
}

func (handler *Handler) Health(writer http.ResponseWriter, req *http.Request) {
	utils.SendJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}
