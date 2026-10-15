package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"switchboard/internal/models"
)

func TestInMemoryStore(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "seller",
	}

	err := store.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}

	err = store.CreateUser(ctx, &models.User{
		Name:  "Dup User",
		Email: "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error creating duplicate user, got nil")
	}

	_, err = store.GetUserByID(ctx, 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for user, got %v", err)
	}

	_, err = store.GetUserByEmail(ctx, "nonexistent@example.com")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	fetchedUser, err := store.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching user by ID: %v", err)
	}
	if fetchedUser.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, fetchedUser.Email)
	}

	listing := &models.Listing{
		Title:       "Test Bike",
		Description: "A nice test bike",
		Price:       150.00,
		Category:    "Bikes",
		SellerID:    user.ID,
		CreatedAt:   time.Now(),
	}

	err = store.CreateListing(ctx, listing)
	if err != nil {
		t.Fatalf("unexpected error creating listing: %v", err)
	}

	allListings, err := store.GetListings(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting listings: %v", err)
	}
	if len(allListings) != 1 {
		t.Errorf("expected 1 listing, got %d", len(allListings))
	}

	_, err = store.GetListingByID(ctx, listing.ID)
	if err != nil {
		t.Fatalf("unexpected error getting listing by ID: %v", err)
	}

	_, err = store.GetListingByID(ctx, 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for listing, got %v", err)
	}

	listing.Title = "Updated Bike"
	err = store.UpdateListing(ctx, listing)
	if err != nil {
		t.Fatalf("unexpected error updating listing: %v", err)
	}

	err = store.UpdateListing(ctx, &models.Listing{ID: 9999})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on update listing, got %v", err)
	}

	err = store.CreateBookmark(ctx, user.ID, listing.ID)
	if err != nil {
		t.Fatalf("unexpected error creating bookmark: %v", err)
	}

	err = store.CreateBookmark(ctx, user.ID, listing.ID)
	if err != nil {
		t.Fatalf("unexpected error creating duplicate bookmark: %v", err)
	}

	bookmarks, err := store.GetBookmarksForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("unexpected error getting bookmarks: %v", err)
	}
	if len(bookmarks) != 1 || bookmarks[0] != listing.ID {
		t.Errorf("expected bookmarks [1], got %v", bookmarks)
	}

	interested, err := store.GetInterestedUsersForListing(ctx, listing.ID)
	if err != nil {
		t.Fatalf("unexpected error getting interested users: %v", err)
	}
	if len(interested) != 1 || interested[0] != user.Email {
		t.Errorf("expected interested users [test@example.com], got %v", interested)
	}

	searchResults, err := store.SearchListings(ctx, "bike", "Bikes", 100, 200, "price", "asc")
	if err != nil {
		t.Fatalf("unexpected error searching listings: %v", err)
	}
	if len(searchResults) != 1 {
		t.Errorf("expected 1 search result, got %d", len(searchResults))
	}

	searchResults, err = store.SearchListings(ctx, "bike", "Bikes", 100, 200, "created_at", "desc")
	if err != nil {
		t.Fatalf("unexpected error searching listings: %v", err)
	}
	if len(searchResults) != 1 {
		t.Errorf("expected 1 search result, got %d", len(searchResults))
	}

	err = store.DeleteBookmark(ctx, user.ID, listing.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting bookmark: %v", err)
	}

	err = store.DeleteListing(ctx, listing.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting listing: %v", err)
	}

	err = store.DeleteListing(ctx, 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on delete listing, got %v", err)
	}
}
