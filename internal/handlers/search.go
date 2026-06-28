package handlers

import (
	"net/http"
	"strconv"

	"switchboard/internal/utils"
)

func (handler *Handler) SearchListings(writer http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()

	query := queryParams.Get("q")
	category := queryParams.Get("category")

	var minPrice float64
	if minPriceStr := queryParams.Get("min_price"); minPriceStr != "" {
		val, err := strconv.ParseFloat(minPriceStr, 64)
		if err == nil {
			minPrice = val
		}
	}

	var maxPrice float64
	if maxPriceStr := queryParams.Get("max_price"); maxPriceStr != "" {
		val, err := strconv.ParseFloat(maxPriceStr, 64)
		if err == nil {
			maxPrice = val
		}
	}

	sortBy := queryParams.Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := queryParams.Get("sort_order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	results, err := handler.Listings.SearchListings(req.Context(), query, category, minPrice, maxPrice, sortBy, sortOrder)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	utils.SendJSON(writer, http.StatusOK, results)
}
