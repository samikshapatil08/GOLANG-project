package db

import (
	"context"
	"time"
)

const createUser = `
INSERT INTO users (name, email, password, role)
VALUES ($1, $2, $3, $4)
RETURNING id, name, email, password, role
`

func (queries *Queries) CreateUser(ctx context.Context, name string, email string, password string, role string) (User, error) {
	row := queries.db.QueryRowContext(ctx, createUser, name, email, password, role)
	var user User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
	)
	return user, err
}

const getUserByEmail = `
SELECT id, name, email, password, role
FROM users
WHERE email = $1
`

func (queries *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := queries.db.QueryRowContext(ctx, getUserByEmail, email)
	var user User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
	)
	return user, err
}

const getUserByID = `
SELECT id, name, email, password, role
FROM users
WHERE id = $1
`

func (queries *Queries) GetUserByID(ctx context.Context, id int32) (User, error) {
	row := queries.db.QueryRowContext(ctx, getUserByID, id)
	var user User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
	)
	return user, err
}

const createListing = `
INSERT INTO listings (title, description, price, category, seller_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, title, description, price, category, seller_id, created_at
`

func (queries *Queries) CreateListing(ctx context.Context, title string, description string, price float64, category string, sellerID int32, createdAt time.Time) (Listing, error) {
	row := queries.db.QueryRowContext(ctx, createListing, title, description, price, category, sellerID, createdAt)
	var listing Listing
	err := row.Scan(
		&listing.ID,
		&listing.Title,
		&listing.Description,
		&listing.Price,
		&listing.Category,
		&listing.SellerID,
		&listing.CreatedAt,
	)
	return listing, err
}

const getListings = `
SELECT id, title, description, price, category, seller_id, created_at
FROM listings
`

func (queries *Queries) GetListings(ctx context.Context) ([]Listing, error) {
	rows, err := queries.db.QueryContext(ctx, getListings)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Listing
	for rows.Next() {
		var listing Listing
		err := rows.Scan(
			&listing.ID,
			&listing.Title,
			&listing.Description,
			&listing.Price,
			&listing.Category,
			&listing.SellerID,
			&listing.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getListingByID = `
SELECT id, title, description, price, category, seller_id, created_at
FROM listings
WHERE id = $1
`

func (queries *Queries) GetListingByID(ctx context.Context, id int32) (Listing, error) {
	row := queries.db.QueryRowContext(ctx, getListingByID, id)
	var listing Listing
	err := row.Scan(
		&listing.ID,
		&listing.Title,
		&listing.Description,
		&listing.Price,
		&listing.Category,
		&listing.SellerID,
		&listing.CreatedAt,
	)
	return listing, err
}

const updateListing = `
UPDATE listings
SET title = $2, description = $3, price = $4, category = $5, seller_id = $6, created_at = $7
WHERE id = $1
RETURNING id, title, description, price, category, seller_id, created_at
`

func (queries *Queries) UpdateListing(ctx context.Context, id int32, title string, description string, price float64, category string, sellerID int32, createdAt time.Time) (Listing, error) {
	row := queries.db.QueryRowContext(ctx, updateListing, id, title, description, price, category, sellerID, createdAt)
	var listing Listing
	err := row.Scan(
		&listing.ID,
		&listing.Title,
		&listing.Description,
		&listing.Price,
		&listing.Category,
		&listing.SellerID,
		&listing.CreatedAt,
	)
	return listing, err
}

const deleteListing = `
DELETE FROM listings
WHERE id = $1
`

func (queries *Queries) DeleteListing(ctx context.Context, id int32) error {
	_, err := queries.db.ExecContext(ctx, deleteListing, id)
	return err
}

const createBookmark = `
INSERT INTO bookmarks (user_id, listing_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

func (queries *Queries) CreateBookmark(ctx context.Context, userID int32, listingID int32) error {
	_, err := queries.db.ExecContext(ctx, createBookmark, userID, listingID)
	return err
}

const deleteBookmark = `
DELETE FROM bookmarks
WHERE user_id = $1 AND listing_id = $2
`

func (queries *Queries) DeleteBookmark(ctx context.Context, userID int32, listingID int32) error {
	_, err := queries.db.ExecContext(ctx, deleteBookmark, userID, listingID)
	return err
}

const getBookmarksForUser = `
SELECT listing_id FROM bookmarks WHERE user_id = $1
`

func (queries *Queries) GetBookmarksForUser(ctx context.Context, userID int32) ([]int32, error) {
	rows, err := queries.db.QueryContext(ctx, getBookmarksForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []int32
	for rows.Next() {
		var listingID int32
		if err := rows.Scan(&listingID); err != nil {
			return nil, err
		}
		items = append(items, listingID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInterestedUsersForListing = `
SELECT users.email
FROM users
JOIN bookmarks ON users.id = bookmarks.user_id
WHERE bookmarks.listing_id = $1
`

func (queries *Queries) GetInterestedUsersForListing(ctx context.Context, listingID int32) ([]string, error) {
	rows, err := queries.db.QueryContext(ctx, getInterestedUsersForListing, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		items = append(items, email)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
