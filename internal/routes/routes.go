package routes

import (
	"github.com/go-chi/chi/v5"

	"switchboard/internal/handlers"
	"switchboard/internal/middleware"
)

func Setup(handler *handlers.Handler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/health", handler.Health)
	router.Get("/listings/sample", handler.SampleListings)
	router.Get("/listings/price-check", handler.PriceCheck)
	router.Get("/listings", handler.GetListings)
	router.Get("/listings/recommended", handler.GetRecommendedListings)
	router.Get("/listings/{id}", handler.GetListingByID)
	router.Post("/listings/{id}/notify-interested", handler.NotifyInterested)
	router.Get("/search", handler.SearchListings)
	router.Post("/api/price-advisor", handler.PriceAdvisor)

	router.Post("/auth/register", handler.Register)
	router.Post("/auth/login", handler.Login)
	router.Post("/auth/refresh", handler.Refresh)
	router.Post("/auth/logout", handler.Logout)

	router.Group(func(subRouter chi.Router) {
		subRouter.Use(middleware.RequireAuth(handler.JWTSecret))
		subRouter.Post("/listings", handler.CreateListing)
		subRouter.Put("/listings/{id}", handler.UpdateListing)
		subRouter.Patch("/listings/{id}", handler.PatchListing)
		subRouter.Delete("/listings/{id}", handler.DeleteListing)
	})

	return router
}
