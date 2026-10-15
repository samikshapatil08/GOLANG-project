package main

import (
	"fmt"
	"sort"
)

type Listing struct {
	Title         string
	Price         float64
	FavoriteCount int
}

func TopN(listings []Listing, limit int) []Listing {
	sort.Slice(listings, func(indexFirst, indexSecond int) bool {
		if listings[indexFirst].FavoriteCount != listings[indexSecond].FavoriteCount {
			return listings[indexFirst].FavoriteCount > listings[indexSecond].FavoriteCount
		}
		return listings[indexFirst].Price < listings[indexSecond].Price
	})

	if limit > len(listings) {
		limit = len(listings)
	}
	if limit < 0 {
		limit = 0
	}

	return listings[:limit]
}

func main() {
	sampleListings := []Listing{
		{Title: "Bicycle", Price: 150.0, FavoriteCount: 5},
		{Title: "Laptop", Price: 800.0, FavoriteCount: 12},
		{Title: "Textbook", Price: 40.0, FavoriteCount: 5},
		{Title: "Desk Lamp", Price: 15.0, FavoriteCount: 2},
		{Title: "Headphones", Price: 100.0, FavoriteCount: 12},
	}

	fmt.Println("Original Listings:")
	for _, item := range sampleListings {
		fmt.Printf("- %s: Price: $%.2f, Favorites: %d\n", item.Title, item.Price, item.FavoriteCount)
	}

	topThree := TopN(sampleListings, 3)

	fmt.Println("\nTop 3 Ranked Listings:")
	for _, item := range topThree {
		fmt.Printf("- %s: Price: $%.2f, Favorites: %d\n", item.Title, item.Price, item.FavoriteCount)
	}
}
