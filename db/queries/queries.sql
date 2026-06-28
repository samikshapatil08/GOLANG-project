-- name: CreateUser :one
INSERT INTO users (name, email, password, role)
VALUES ($1, $2, $3, $4)
RETURNING id, name, email, password, role;

-- name: GetUserByEmail :one
SELECT id, name, email, password, role
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, name, email, password, role
FROM users
WHERE id = $1;

-- name: CreateListing :one
INSERT INTO listings (title, description, price, category, seller_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, title, description, price, category, seller_id, created_at;

-- name: GetListings :many
SELECT id, title, description, price, category, seller_id, created_at
FROM listings;

-- name: GetListingByID :one
SELECT id, title, description, price, category, seller_id, created_at
FROM listings
WHERE id = $1;

-- name: UpdateListing :one
UPDATE listings
SET title = $2, description = $3, price = $4, category = $5, seller_id = $6, created_at = $7
WHERE id = $1
RETURNING id, title, description, price, category, seller_id, created_at;

-- name: DeleteListing :exec
DELETE FROM listings
WHERE id = $1;

-- name: CreateBookmark :exec
INSERT INTO bookmarks (user_id, listing_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteBookmark :exec
DELETE FROM bookmarks
WHERE user_id = $1 AND listing_id = $2;

-- name: GetBookmarksForUser :many
SELECT listing_id FROM bookmarks WHERE user_id = $1;

-- name: GetInterestedUsersForListing :many
SELECT users.email
FROM users
JOIN bookmarks ON users.id = bookmarks.user_id
WHERE bookmarks.listing_id = $1;
