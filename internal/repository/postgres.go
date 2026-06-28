package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"switchboard/internal/models"
	"switchboard/internal/repository/db"
)

type PostgresStore struct {
	queries *db.Queries
	rawDB   *sql.DB
}

func NewPostgresStore(rawDB *sql.DB) *PostgresStore {
	return &PostgresStore{
		queries: db.New(rawDB),
		rawDB:   rawDB,
	}
}

func (store *PostgresStore) CreateListing(ctx context.Context, listing *models.Listing) error {
	result, err := store.queries.CreateListing(
		ctx,
		listing.Title,
		listing.Description,
		listing.Price,
		listing.Category,
		int32(listing.SellerID),
		listing.CreatedAt,
	)
	if err != nil {
		return err
	}

	listing.ID = int(result.ID)
	listing.CreatedAt = result.CreatedAt
	return nil
}

func (store *PostgresStore) GetListings(ctx context.Context) ([]models.Listing, error) {
	dbListings, err := store.queries.GetListings(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]models.Listing, 0, len(dbListings))
	for _, item := range dbListings {
		results = append(results, models.Listing{
			ID:          int(item.ID),
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			Category:    item.Category,
			SellerID:    int(item.SellerID),
			CreatedAt:   item.CreatedAt,
		})
	}
	return results, nil
}

func (store *PostgresStore) GetListingByID(ctx context.Context, id int) (*models.Listing, error) {
	item, err := store.queries.GetListingByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &models.Listing{
		ID:          int(item.ID),
		Title:       item.Title,
		Description: item.Description,
		Price:       item.Price,
		Category:    item.Category,
		SellerID:    int(item.SellerID),
		CreatedAt:   item.CreatedAt,
	}, nil
}

func (store *PostgresStore) UpdateListing(ctx context.Context, listing *models.Listing) error {
	_, err := store.queries.UpdateListing(
		ctx,
		int32(listing.ID),
		listing.Title,
		listing.Description,
		listing.Price,
		listing.Category,
		int32(listing.SellerID),
		listing.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (store *PostgresStore) DeleteListing(ctx context.Context, id int) error {
	err := store.queries.DeleteListing(ctx, int32(id))
	if err != nil {
		return err
	}
	return nil
}

func (store *PostgresStore) CreateUser(ctx context.Context, user *models.User) error {
	result, err := store.queries.CreateUser(
		ctx,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
	)
	if err != nil {
		return err
	}

	user.ID = int(result.ID)
	return nil
}

func (store *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	result, err := store.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &models.User{
		ID:       int(result.ID),
		Name:     result.Name,
		Email:    result.Email,
		Password: result.Password,
		Role:     result.Role,
	}, nil
}

func (store *PostgresStore) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	result, err := store.queries.GetUserByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &models.User{
		ID:       int(result.ID),
		Name:     result.Name,
		Email:    result.Email,
		Password: result.Password,
		Role:     result.Role,
	}, nil
}

func (store *PostgresStore) CreateBookmark(ctx context.Context, userID int, listingID int) error {
	return store.queries.CreateBookmark(ctx, int32(userID), int32(listingID))
}

func (store *PostgresStore) DeleteBookmark(ctx context.Context, userID int, listingID int) error {
	return store.queries.DeleteBookmark(ctx, int32(userID), int32(listingID))
}

func (store *PostgresStore) GetBookmarksForUser(ctx context.Context, userID int) ([]int, error) {
	ids, err := store.queries.GetBookmarksForUser(ctx, int32(userID))
	if err != nil {
		return nil, err
	}

	results := make([]int, 0, len(ids))
	for _, id := range ids {
		results = append(results, int(id))
	}
	return results, nil
}

func (store *PostgresStore) GetInterestedUsersForListing(ctx context.Context, listingID int) ([]string, error) {
	return store.queries.GetInterestedUsersForListing(ctx, int32(listingID))
}

func (store *PostgresStore) SearchListings(ctx context.Context, query string, category string, minPrice float64, maxPrice float64, sortBy string, sortOrder string) ([]models.Listing, error) {
	baseQuery := `SELECT id, title, description, price, category, seller_id, created_at FROM listings WHERE 1=1`
	var args []interface{}
	placeholderIndex := 1

	if query != "" {
		baseQuery += fmt.Sprintf(` AND to_tsvector('english', title || ' ' || description) @@ websearch_to_tsquery('english', $%d)`, placeholderIndex)
		args = append(args, query)
		placeholderIndex++
	}

	if category != "" {
		baseQuery += fmt.Sprintf(` AND category = $%d`, placeholderIndex)
		args = append(args, category)
		placeholderIndex++
	}

	if minPrice > 0 {
		baseQuery += fmt.Sprintf(` AND price >= $%d`, placeholderIndex)
		args = append(args, minPrice)
		placeholderIndex++
	}

	if maxPrice > 0 {
		baseQuery += fmt.Sprintf(` AND price <= $%d`, placeholderIndex)
		args = append(args, maxPrice)
		placeholderIndex++
	}

	sortCol := "created_at"
	if sortBy == "price" {
		sortCol = "price"
	}

	direction := "DESC"
	if strings.ToLower(sortOrder) == "asc" {
		direction = "ASC"
	}

	baseQuery += fmt.Sprintf(` ORDER BY %s %s`, sortCol, direction)

	rows, err := store.rawDB.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Listing
	for rows.Next() {
		var item db.Listing
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Price,
			&item.Category,
			&item.SellerID,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, models.Listing{
			ID:          int(item.ID),
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			Category:    item.Category,
			SellerID:    int(item.SellerID),
			CreatedAt:   item.CreatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
