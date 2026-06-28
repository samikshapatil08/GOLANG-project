# Switchboard

Switchboard is a campus marketplace and service-booking exchange API built in Go. It enables students to buy, sell, bookmark listings, request AI pricing recommendations, and search items.

## Architecture

The project is structured as follows:

* cmd/server: Application entry point and server initialization.
* cmd/cli: Standalone command-line utilities.
* internal/config: Environment configuration management.
* internal/models: Data representations.
* internal/handlers: HTTP endpoint handlers.
* internal/middleware: Authentication and logging middlewares.
* internal/repository: Data layer abstractions with PostgreSQL and in-memory implementations.
* internal/routes: API router setup and route definition.
* internal/utils: Shared helpers for auth and JSON responses.

## Core Features

1. User Authentication: Registration and login using bcrypt password hashing and dual-token JWT flow. Access tokens are passed via headers; refresh tokens are set via secure, HTTP-only cookies.
2. Listings Management: Standard CRUD endpoints for listings. It includes ownership validation ensuring only the seller can edit or delete their listings.
3. Bookmarks and Notifications: Users can bookmark/favorite listings. When a notification request is dispatched, a concurrent worker pool notifies all interested bookmark owners.
4. Recommendations: Combines pricing, distance, and popularity data in parallel using errgroup with context timeouts.
5. AI Price Advisor: Integrates Google GenAI to suggest fair listing prices, improved titles, and descriptions.
6. Full-Text Search: Advanced search utilizing PostgreSQL full-text search syntax, filters by categories, price ranges, and dynamic sorting.

## Standalone CLI Utilities

* cmd/cli/handle_gen: Reads a student name from standard input, processes Unicode characters, and appends a random three-digit number to generate a secure campus handle.
* cmd/cli/ranker: Demonstrates Go sorting patterns by sorting a slice of listings by favorites count (descending) and price (ascending) using sort.Slice and closure scopes.

## Go Concepts Demonstrated

* Concurrency: Uses goroutines, errgroup for fan-out API service orchestration, worker pools with channel dispatch, and sync.WaitGroup for clean shutdown management.
* Interfaces and Mocking: Implements repository interfaces allowing the database layer to fall back to an in-memory repository if PostgreSQL is unavailable. This structure is heavily utilized to run mocks during handler tests.
* Context propagation: Context timeouts and cancellations are passed down to external API requests and database queries.
* Defensive Input Validation: Built-in validator patterns protect model fields across write requests.

## Getting Started

1. Environment Configuration:
   Copy the example environment template:
   cp .env.example .env

2. Run Local Server:
   go run cmd/server/main.go

3. Running Tests:
   go test -v -race ./...

4. Run via Docker Compose:
   docker compose up --build
