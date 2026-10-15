package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"

	"switchboard/internal/models"
	"switchboard/internal/repository"
	"switchboard/internal/utils"
)

type RecommendationData struct {
	SuggestedPriceRange string  `json:"suggested_price_range"`
	DistanceMiles       float64 `json:"distance_miles"`
	PopularityScore     int     `json:"popularity_score"`
}

type RecommendedListingResponse struct {
	models.Listing
	Recommendation RecommendationData `json:"recommendation"`
}

func (handler *Handler) mockPricingService(ctx context.Context, id int, price float64) (string, error) {
	select {
	case <-time.After(30 * time.Millisecond):
		return fmt.Sprintf("$%.2f - $%.2f", price*0.9, price*1.1), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (handler *Handler) mockDistanceService(ctx context.Context, id int) (float64, error) {
	select {
	case <-time.After(40 * time.Millisecond):
		return 1.5, nil
	case <-ctx.Done():
		return 0.0, ctx.Err()
	}
}

func (handler *Handler) mockPopularityService(ctx context.Context, id int) (int, error) {
	select {
	case <-time.After(20 * time.Millisecond):
		return 85, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (handler *Handler) fetchRecommendation(ctx context.Context, item models.Listing) (RecommendationData, error) {
	var details RecommendationData
	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		priceRange, err := handler.mockPricingService(groupCtx, item.ID, item.Price)
		if err != nil {
			return err
		}
		details.SuggestedPriceRange = priceRange
		return nil
	})

	group.Go(func() error {
		dist, err := handler.mockDistanceService(groupCtx, item.ID)
		if err != nil {
			return err
		}
		details.DistanceMiles = dist
		return nil
	})

	group.Go(func() error {
		pop, err := handler.mockPopularityService(groupCtx, item.ID)
		if err != nil {
			return err
		}
		details.PopularityScore = pop
		return nil
	})

	err := group.Wait()
	return details, err
}

func (handler *Handler) GetRecommendedListings(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	listings, err := handler.Listings.GetListings(ctx)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to get listings")
		return
	}

	results := make([]RecommendedListingResponse, 0, len(listings))
	for _, item := range listings {
		rec, err := handler.fetchRecommendation(ctx, item)
		if err != nil {
			utils.SendError(writer, http.StatusInternalServerError, "failed to fetch recommendation: "+err.Error())
			return
		}
		results = append(results, RecommendedListingResponse{
			Listing:        item,
			Recommendation: rec,
		})
	}

	utils.SendJSON(writer, http.StatusOK, results)
}

func (handler *Handler) NotifyInterested(writer http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid listing id")
		return
	}

	_, err = handler.Listings.GetListingByID(req.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.SendError(writer, http.StatusNotFound, "listing not found")
			return
		}
		utils.SendError(writer, http.StatusInternalServerError, "failed to fetch listing")
		return
	}

	emails, err := handler.Users.GetInterestedUsersForListing(req.Context(), id)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to get interested users")
		return
	}

	if len(emails) == 0 {
		utils.SendJSON(writer, http.StatusOK, map[string]string{"message": "no interested users to notify"})
		return
	}

	jobsChan := make(chan string, len(emails))
	for _, email := range emails {
		jobsChan <- email
	}
	close(jobsChan)

	workerCount := 3
	if len(emails) < workerCount {
		workerCount = len(emails)
	}

	var waitGroup sync.WaitGroup
	for index := 0; index < workerCount; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for email := range jobsChan {
				slog.Info("notification sent successfully", "email", email, "listing_id", id)
			}
		}()
	}

	waitGroup.Wait()

	utils.SendJSON(writer, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("notifications sent to %d users", len(emails)),
	})
}
