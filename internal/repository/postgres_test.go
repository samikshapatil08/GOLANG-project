package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"switchboard/internal/models"
)

func TestPostgresStore_GetListings(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "title", "description", "price", "category", "seller_id", "created_at"}).
		AddRow(1, "Test Item", "Desc", 10.5, "Cat", 2, time.Now())

	mock.ExpectQuery(`SELECT id, title, description, price, category, seller_id, created_at`).
		WillReturnRows(expectedRows)

	ctx := context.Background()
	results, err := store.GetListings(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPostgresStore_GetListingByID(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "title", "description", "price", "category", "seller_id", "created_at"}).
		AddRow(1, "Test Item", "Desc", 10.5, "Cat", 2, time.Now())

	mock.ExpectQuery(`SELECT id, title, description, price, category, seller_id, created_at FROM listings WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(expectedRows)

	ctx := context.Background()
	listing, err := store.GetListingByID(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listing.Title != "Test Item" {
		t.Errorf("expected title 'Test Item', got %s", listing.Title)
	}
}

func TestPostgresStore_GetListingByID_NotFound(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectQuery(`SELECT id, title, description, price, category, seller_id, created_at FROM listings WHERE id = \$1`).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = store.GetListingByID(ctx, 99)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresStore_CreateListing(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "title", "description", "price", "category", "seller_id", "created_at"}).
		AddRow(1, "Test Item", "Desc", 10.5, "Cat", 2, time.Now())

	mock.ExpectQuery(`INSERT INTO listings`).
		WithArgs("Test Item", "Desc", 10.5, "Cat", 2, sqlmock.AnyArg()).
		WillReturnRows(expectedRows)

	listing := &models.Listing{
		Title:       "Test Item",
		Description: "Desc",
		Price:       10.5,
		Category:    "Cat",
		SellerID:    2,
	}

	ctx := context.Background()
	err = store.CreateListing(ctx, listing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listing.ID != 1 {
		t.Errorf("expected created listing ID 1, got %d", listing.ID)
	}
}

func TestPostgresStore_UpdateListing(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "title", "description", "price", "category", "seller_id", "created_at"}).
		AddRow(1, "Updated Title", "Desc", 10.5, "Cat", 2, time.Now())

	mock.ExpectQuery(`UPDATE listings`).
		WithArgs(1, "Updated Title", "Desc", 10.5, "Cat", 2, sqlmock.AnyArg()).
		WillReturnRows(expectedRows)

	listing := &models.Listing{
		ID:          1,
		Title:       "Updated Title",
		Description: "Desc",
		Price:       10.5,
		Category:    "Cat",
		SellerID:    2,
	}

	ctx := context.Background()
	err = store.UpdateListing(ctx, listing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresStore_UpdateListing_NotFound(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectQuery(`UPDATE listings`).
		WithArgs(99, "Updated Title", "Desc", 10.5, "Cat", 2, sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	listing := &models.Listing{
		ID:          99,
		Title:       "Updated Title",
		Description: "Desc",
		Price:       10.5,
		Category:    "Cat",
		SellerID:    2,
	}

	ctx := context.Background()
	err = store.UpdateListing(ctx, listing)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresStore_DeleteListing(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectExec(`DELETE FROM listings`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err = store.DeleteListing(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresStore_DeleteListing_NotFound(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectExec(`DELETE FROM listings`).
		WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 0))

	ctx := context.Background()
	err = store.DeleteListing(ctx, 99)
	if err != nil {
		t.Errorf("expected nil error for nonexistent delete, got %v", err)
	}
}

func TestPostgresStore_CreateUser(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role"}).
		AddRow(1, "User Name", "user@example.com", "pass", "seller")

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("User Name", "user@example.com", "pass", "seller").
		WillReturnRows(expectedRows)

	user := &models.User{
		Name:     "User Name",
		Email:    "user@example.com",
		Password: "pass",
		Role:     "seller",
	}

	ctx := context.Background()
	err = store.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresStore_GetUserByEmail(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role"}).
		AddRow(1, "User Name", "user@example.com", "pass", "seller")

	mock.ExpectQuery(`SELECT id, name, email, password, role FROM users WHERE email = \$1`).
		WithArgs("user@example.com").
		WillReturnRows(expectedRows)

	ctx := context.Background()
	user, err := store.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Name != "User Name" {
		t.Errorf("expected name 'User Name', got %s", user.Name)
	}
}

func TestPostgresStore_GetUserByEmail_NotFound(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectQuery(`SELECT id, name, email, password, role FROM users WHERE email = \$1`).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = store.GetUserByEmail(ctx, "missing@example.com")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresStore_GetUserByID(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role"}).
		AddRow(1, "User Name", "user@example.com", "pass", "seller")

	mock.ExpectQuery(`SELECT id, name, email, password, role FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(expectedRows)

	ctx := context.Background()
	user, err := store.GetUserByID(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Name != "User Name" {
		t.Errorf("expected name 'User Name', got %s", user.Name)
	}
}

func TestPostgresStore_GetUserByID_NotFound(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectQuery(`SELECT id, name, email, password, role FROM users WHERE id = \$1`).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = store.GetUserByID(ctx, 99)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresStore_CreateBookmark(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectExec(`INSERT INTO bookmarks`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err = store.CreateBookmark(ctx, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresStore_DeleteBookmark(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	mock.ExpectExec(`DELETE FROM bookmarks`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err = store.DeleteBookmark(ctx, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresStore_GetBookmarksForUser(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"listing_id"}).
		AddRow(2).
		AddRow(3)

	mock.ExpectQuery(`SELECT listing_id FROM bookmarks WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(expectedRows)

	ctx := context.Background()
	results, err := store.GetBookmarksForUser(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 bookmarks, got %d", len(results))
	}
}

func TestPostgresStore_GetInterestedUsersForListing(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"email"}).
		AddRow("interested@example.com")

	mock.ExpectQuery(`SELECT users.email FROM users JOIN bookmarks ON users.id = bookmarks.user_id WHERE bookmarks.listing_id = \$1`).
		WithArgs(1).
		WillReturnRows(expectedRows)

	ctx := context.Background()
	results, err := store.GetInterestedUsersForListing(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 || results[0] != "interested@example.com" {
		t.Errorf("expected ['interested@example.com'], got %v", results)
	}
}

func TestPostgresStore_SearchListings(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer dbConn.Close()

	store := NewPostgresStore(dbConn)

	expectedRows := sqlmock.NewRows([]string{"id", "title", "description", "price", "category", "seller_id", "created_at"}).
		AddRow(1, "Search Item", "Desc", 10.5, "Cat", 2, time.Now())

	mock.ExpectQuery(`SELECT id, title, description, price, category, seller_id, created_at FROM listings WHERE 1=1`).
		WithArgs("searchquery").
		WillReturnRows(expectedRows)

	ctx := context.Background()
	results, err := store.SearchListings(ctx, "searchquery", "", 0, 0, "price", "asc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 || results[0].Title != "Search Item" {
		t.Errorf("expected 1 result, got %v", results)
	}
}
