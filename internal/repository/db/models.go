package db

import (
	"time"
)

type User struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Listing struct {
	ID          int32     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	SellerID    int32     `json:"seller_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Bookmark struct {
	UserID    int32 `json:"user_id"`
	ListingID int32 `json:"listing_id"`
}
