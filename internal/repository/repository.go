package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"switchboard/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
)

type ListingRepository interface {
	CreateListing(ctx context.Context, listing *models.Listing) error
	GetListings(ctx context.Context) ([]models.Listing, error)
	GetListingByID(ctx context.Context, id int) (*models.Listing, error)
	UpdateListing(ctx context.Context, listing *models.Listing) error
	DeleteListing(ctx context.Context, id int) error
	SearchListings(ctx context.Context, query string, category string, minPrice float64, maxPrice float64, sortBy string, sortOrder string) ([]models.Listing, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	CreateBookmark(ctx context.Context, userID int, listingID int) error
	DeleteBookmark(ctx context.Context, userID int, listingID int) error
	GetBookmarksForUser(ctx context.Context, userID int) ([]int, error)
	GetInterestedUsersForListing(ctx context.Context, listingID int) ([]string, error)
}

type InMemoryStore struct {
	listingsLock sync.RWMutex
	usersLock    sync.RWMutex
	listings     map[int]models.Listing
	users        map[int]models.User
	bookmarks    map[int][]int
	nextListID   int
	nextUserID   int
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		listings:   make(map[int]models.Listing),
		users:      make(map[int]models.User),
		bookmarks:  make(map[int][]int),
		nextListID: 1,
		nextUserID: 1,
	}
}

func (store *InMemoryStore) CreateListing(ctx context.Context, listing *models.Listing) error {
	store.listingsLock.Lock()
	defer store.listingsLock.Unlock()

	listing.ID = store.nextListID
	listing.CreatedAt = time.Now()
	store.listings[listing.ID] = *listing
	store.nextListID++
	return nil
}

func (store *InMemoryStore) GetListings(ctx context.Context) ([]models.Listing, error) {
	store.listingsLock.RLock()
	defer store.listingsLock.RUnlock()

	results := make([]models.Listing, 0, len(store.listings))
	for _, item := range store.listings {
		results = append(results, item)
	}
	return results, nil
}

func (store *InMemoryStore) GetListingByID(ctx context.Context, id int) (*models.Listing, error) {
	store.listingsLock.RLock()
	defer store.listingsLock.RUnlock()

	item, exists := store.listings[id]
	if !exists {
		return nil, ErrNotFound
	}
	return &item, nil
}

func (store *InMemoryStore) UpdateListing(ctx context.Context, listing *models.Listing) error {
	store.listingsLock.Lock()
	defer store.listingsLock.Unlock()

	_, exists := store.listings[listing.ID]
	if !exists {
		return ErrNotFound
	}
	store.listings[listing.ID] = *listing
	return nil
}

func (store *InMemoryStore) DeleteListing(ctx context.Context, id int) error {
	store.listingsLock.Lock()
	defer store.listingsLock.Unlock()

	_, exists := store.listings[id]
	if !exists {
		return ErrNotFound
	}
	delete(store.listings, id)
	return nil
}

func (store *InMemoryStore) CreateUser(ctx context.Context, user *models.User) error {
	store.usersLock.Lock()
	defer store.usersLock.Unlock()

	for _, existing := range store.users {
		if existing.Email == user.Email {
			return errors.New("user already exists")
		}
	}

	user.ID = store.nextUserID
	store.users[user.ID] = *user
	store.nextUserID++
	return nil
}

func (store *InMemoryStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	store.usersLock.RLock()
	defer store.usersLock.RUnlock()

	for _, user := range store.users {
		if user.Email == email {
			userCopy := user
			return &userCopy, nil
		}
	}
	return nil, ErrNotFound
}

func (store *InMemoryStore) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	store.usersLock.RLock()
	defer store.usersLock.RUnlock()

	user, exists := store.users[id]
	if !exists {
		return nil, ErrNotFound
	}
	return &user, nil
}

func (store *InMemoryStore) CreateBookmark(ctx context.Context, userID int, listingID int) error {
	store.usersLock.Lock()
	defer store.usersLock.Unlock()

	list := store.bookmarks[userID]
	for _, id := range list {
		if id == listingID {
			return nil
		}
	}
	store.bookmarks[userID] = append(list, listingID)
	return nil
}

func (store *InMemoryStore) DeleteBookmark(ctx context.Context, userID int, listingID int) error {
	store.usersLock.Lock()
	defer store.usersLock.Unlock()

	list := store.bookmarks[userID]
	for index, id := range list {
		if id == listingID {
			store.bookmarks[userID] = append(list[:index], list[index+1:]...)
			return nil
		}
	}
	return nil
}

func (store *InMemoryStore) GetBookmarksForUser(ctx context.Context, userID int) ([]int, error) {
	store.usersLock.RLock()
	defer store.usersLock.RUnlock()

	return store.bookmarks[userID], nil
}

func (store *InMemoryStore) GetInterestedUsersForListing(ctx context.Context, listingID int) ([]string, error) {
	store.usersLock.RLock()
	defer store.usersLock.RUnlock()

	var emails []string
	for userID, listIDs := range store.bookmarks {
		for _, id := range listIDs {
			if id == listingID {
				if user, exists := store.users[userID]; exists {
					emails = append(emails, user.Email)
				}
				break
			}
		}
	}
	return emails, nil
}

func (store *InMemoryStore) SearchListings(ctx context.Context, query string, category string, minPrice float64, maxPrice float64, sortBy string, sortOrder string) ([]models.Listing, error) {
	store.listingsLock.RLock()
	defer store.listingsLock.RUnlock()

	var results []models.Listing
	for _, item := range store.listings {
		if query != "" {
			queryLower := strings.ToLower(query)
			titleLower := strings.ToLower(item.Title)
			descLower := strings.ToLower(item.Description)
			if !strings.Contains(titleLower, queryLower) && !strings.Contains(descLower, queryLower) {
				continue
			}
		}

		if category != "" && strings.ToLower(item.Category) != strings.ToLower(category) {
			continue
		}

		if minPrice > 0 && item.Price < minPrice {
			continue
		}

		if maxPrice > 0 && item.Price > maxPrice {
			continue
		}

		results = append(results, item)
	}

	sort.Slice(results, func(i, j int) bool {
		isLess := false
		if sortBy == "price" {
			isLess = results[i].Price < results[j].Price
		} else {
			isLess = results[i].CreatedAt.Before(results[j].CreatedAt)
		}

		if strings.ToLower(sortOrder) == "desc" {
			return !isLess
		}
		return isLess
	})

	return results, nil
}
