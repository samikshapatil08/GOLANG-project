package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"switchboard/internal/middleware"
	"switchboard/internal/models"
	"switchboard/internal/repository"
	"switchboard/internal/utils"
)

type SampleListing struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type PriceCheckResponse struct {
	Item                  string `json:"item"`
	SuggestedCurrencyNote string `json:"suggested_currency_note"`
	Source                string `json:"source"`
}

type ExchangeRatesResponse struct {
	Result   string             `json:"result"`
	Provider string             `json:"provider"`
	BaseCode string             `json:"base_code"`
	Rates    map[string]float64 `json:"rates"`
}

func (handler *Handler) SampleListings(writer http.ResponseWriter, req *http.Request) {
	samples := []SampleListing{
		{ID: 1, Title: "Bicycle", Price: 120.50, Category: "Transport"},
		{ID: 2, Title: "Textbook", Price: 45.00, Category: "Education"},
		{ID: 3, Title: "Desk Lamp", Price: 15.00, Category: "Furniture"},
		{ID: 4, Title: "Headphones", Price: 85.00, Category: "Electronics"},
	}
	utils.SendJSON(writer, http.StatusOK, samples)
}

func (handler *Handler) PriceCheck(writer http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	if item == "" {
		utils.SendError(writer, http.StatusBadRequest, "item query parameter is required")
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	externalReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://open.er-api.com/v6/latest/USD", nil)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to create request")
		return
	}

	resp, err := client.Do(externalReq)
	if err != nil {
		utils.SendError(writer, http.StatusBadGateway, "failed to contact upstream API")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.SendError(writer, http.StatusBadGateway, "upstream API returned non-200 status")
		return
	}

	var rateData ExchangeRatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&rateData); err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to decode upstream response")
		return
	}

	eurRate := rateData.Rates["EUR"]
	gbpRate := rateData.Rates["GBP"]
	note := fmt.Sprintf("Standard price check for %s. Current USD base rates: EUR %.4f, GBP %.4f", item, eurRate, gbpRate)

	result := PriceCheckResponse{
		Item:                  item,
		SuggestedCurrencyNote: note,
		Source:                rateData.Provider,
	}

	utils.SendJSON(writer, http.StatusOK, result)
}

type CreateListingRequest struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Category    string  `json:"category" validate:"required"`
}

func (handler *Handler) CreateListing(writer http.ResponseWriter, req *http.Request) {
	userID, ok := middleware.GetUserID(req.Context())
	if !ok {
		utils.SendError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}

	var reqBody CreateListingRequest
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := handler.validate.Struct(reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, err.Error())
		return
	}

	listing := &models.Listing{
		Title:       reqBody.Title,
		Description: reqBody.Description,
		Price:       reqBody.Price,
		Category:    reqBody.Category,
		SellerID:    userID,
	}

	err := handler.Listings.CreateListing(req.Context(), listing)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to create listing")
		return
	}

	utils.SendJSON(writer, http.StatusCreated, listing)
}

func (handler *Handler) GetListings(writer http.ResponseWriter, req *http.Request) {
	items, err := handler.Listings.GetListings(req.Context())
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to get listings")
		return
	}
	utils.SendJSON(writer, http.StatusOK, items)
}

func (handler *Handler) GetListingByID(writer http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid listing id")
		return
	}

	item, err := handler.Listings.GetListingByID(req.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.SendError(writer, http.StatusNotFound, "listing not found")
			return
		}
		utils.SendError(writer, http.StatusInternalServerError, "failed to get listing")
		return
	}
	utils.SendJSON(writer, http.StatusOK, item)
}

type UpdateListingRequest struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Category    string  `json:"category" validate:"required"`
}

func (handler *Handler) UpdateListing(writer http.ResponseWriter, req *http.Request) {
	userID, ok := middleware.GetUserID(req.Context())
	if !ok {
		utils.SendError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid listing id")
		return
	}

	item, err := handler.Listings.GetListingByID(req.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.SendError(writer, http.StatusNotFound, "listing not found")
			return
		}
		utils.SendError(writer, http.StatusInternalServerError, "failed to fetch listing")
		return
	}

	if item.SellerID != userID {
		utils.SendError(writer, http.StatusForbidden, "forbidden: you do not own this listing")
		return
	}

	var reqBody UpdateListingRequest
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := handler.validate.Struct(reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, err.Error())
		return
	}

	item.Title = reqBody.Title
	item.Description = reqBody.Description
	item.Price = reqBody.Price
	item.Category = reqBody.Category

	err = handler.Listings.UpdateListing(req.Context(), item)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to update listing")
		return
	}

	utils.SendJSON(writer, http.StatusOK, item)
}

type PatchListingRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,gt=0"`
	Category    *string  `json:"category"`
}

func (handler *Handler) PatchListing(writer http.ResponseWriter, req *http.Request) {
	userID, ok := middleware.GetUserID(req.Context())
	if !ok {
		utils.SendError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid listing id")
		return
	}

	item, err := handler.Listings.GetListingByID(req.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.SendError(writer, http.StatusNotFound, "listing not found")
			return
		}
		utils.SendError(writer, http.StatusInternalServerError, "failed to fetch listing")
		return
	}

	if item.SellerID != userID {
		utils.SendError(writer, http.StatusForbidden, "forbidden: you do not own this listing")
		return
	}

	var reqBody PatchListingRequest
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := handler.validate.Struct(reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, err.Error())
		return
	}

	if reqBody.Title != nil {
		item.Title = *reqBody.Title
	}
	if reqBody.Description != nil {
		item.Description = *reqBody.Description
	}
	if reqBody.Price != nil {
		item.Price = *reqBody.Price
	}
	if reqBody.Category != nil {
		item.Category = *reqBody.Category
	}

	err = handler.Listings.UpdateListing(req.Context(), item)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to patch listing")
		return
	}

	utils.SendJSON(writer, http.StatusOK, item)
}

func (handler *Handler) DeleteListing(writer http.ResponseWriter, req *http.Request) {
	userID, ok := middleware.GetUserID(req.Context())
	if !ok {
		utils.SendError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid listing id")
		return
	}

	item, err := handler.Listings.GetListingByID(req.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.SendError(writer, http.StatusNotFound, "listing not found")
			return
		}
		utils.SendError(writer, http.StatusInternalServerError, "failed to fetch listing")
		return
	}

	if item.SellerID != userID {
		utils.SendError(writer, http.StatusForbidden, "forbidden: you do not own this listing")
		return
	}

	err = handler.Listings.DeleteListing(req.Context(), id)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to delete listing")
		return
	}

	utils.SendJSON(writer, http.StatusOK, map[string]string{"message": "listing deleted successfully"})
}
