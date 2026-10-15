package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"switchboard/internal/middleware"
	"switchboard/internal/repository"
	"switchboard/internal/utils"
)

func TestHandlers(t *testing.T) {
	store := repository.NewInMemoryStore()
	secret := "jwtsecretkey123"
	handler := NewHandler(store, store, secret)

	t.Run("Health Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		recorder := httptest.NewRecorder()

		handler.Health(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Sample Listings", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings/sample", nil)
		recorder := httptest.NewRecorder()

		handler.SampleListings(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Register User", func(t *testing.T) {
		reqBody := `{"name":"John Doe","email":"john@example.com","password":"mypassword","role":"seller"}`
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.Register(recorder, req)

		if recorder.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d. Body: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("Register User Duplicate", func(t *testing.T) {
		reqBody := `{"name":"John Doe","email":"john@example.com","password":"mypassword","role":"seller"}`
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.Register(recorder, req)

		if recorder.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", recorder.Code)
		}
	})

	t.Run("Register User Bad Input", func(t *testing.T) {
		reqBody := `{"name":"","email":"bad-email","password":"short","role":""}`
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.Register(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Register User Bad JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString("invalid json"))
		recorder := httptest.NewRecorder()

		handler.Register(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Login User", func(t *testing.T) {
		reqBody := `{"email":"john@example.com","password":"mypassword"}`
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.Login(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Login User Bad Credentials", func(t *testing.T) {
		reqBody := `{"email":"john@example.com","password":"wrongpassword"}`
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.Login(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", recorder.Code)
		}
	})

	t.Run("Login User Bad JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString("invalid json"))
		recorder := httptest.NewRecorder()

		handler.Login(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Login User Not Found", func(t *testing.T) {
		reqBody := `{"email":"missing@example.com","password":"mypassword"}`
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.Login(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", recorder.Code)
		}
	})

	t.Run("Refresh Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/refresh", nil)
		cookieVal, err := utils.GenerateRefreshToken(1, secret)
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: cookieVal,
		})
		recorder := httptest.NewRecorder()

		handler.Refresh(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("Refresh Token Missing Cookie", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/refresh", nil)
		recorder := httptest.NewRecorder()

		handler.Refresh(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", recorder.Code)
		}
	})

	t.Run("Refresh Token Invalid Signature", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/refresh", nil)
		cookieVal, err := utils.GenerateRefreshToken(1, "wrongsecret")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: cookieVal,
		})
		recorder := httptest.NewRecorder()

		handler.Refresh(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", recorder.Code)
		}
	})

	t.Run("Create Listing", func(t *testing.T) {
		reqBody := `{"title":"Vintage Watch","description":"Old watch","price":299.99,"category":"Watches"}`
		req := httptest.NewRequest("POST", "/listings", bytes.NewBufferString(reqBody))

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.CreateListing(recorder, req)

		if recorder.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d. Body: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("Create Listing Bad JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/listings", bytes.NewBufferString("invalid json"))

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.CreateListing(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Create Listing Unauthorized", func(t *testing.T) {
		reqBody := `{"title":"Vintage Watch","description":"Old watch","price":299.99,"category":"Watches"}`
		req := httptest.NewRequest("POST", "/listings", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.CreateListing(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", recorder.Code)
		}
	})

	t.Run("Get Listings", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings", nil)
		recorder := httptest.NewRecorder()

		handler.GetListings(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Get Listing By ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings/1", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetListingByID(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Get Listing By ID Bad ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings/invalid", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "invalid")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetListingByID(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Get Listing By ID Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings/999", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "999")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetListingByID(recorder, req)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})

	t.Run("Update Listing", func(t *testing.T) {
		reqBody := `{"title":"New Watch","description":"Newer watch","price":399.99,"category":"Watches"}`
		req := httptest.NewRequest("PUT", "/listings/1", bytes.NewBufferString(reqBody))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.UpdateListing(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("Update Listing Bad ID", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/listings/invalid", bytes.NewBufferString("{}"))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "invalid")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.UpdateListing(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Update Listing Ownership Forbidden", func(t *testing.T) {
		reqBody := `{"title":"Forbidden Watch","description":"Newer watch","price":399.99,"category":"Watches"}`
		req := httptest.NewRequest("PUT", "/listings/1", bytes.NewBufferString(reqBody))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 2)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.UpdateListing(recorder, req)

		if recorder.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", recorder.Code)
		}
	})

	t.Run("Update Listing Not Found", func(t *testing.T) {
		reqBody := `{"title":"New Watch","description":"Newer watch","price":399.99,"category":"Watches"}`
		req := httptest.NewRequest("PUT", "/listings/999", bytes.NewBufferString(reqBody))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "999")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.UpdateListing(recorder, req)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})

	t.Run("Patch Listing", func(t *testing.T) {
		reqBody := `{"title":"Patched Watch"}`
		req := httptest.NewRequest("PATCH", "/listings/1", bytes.NewBufferString(reqBody))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.PatchListing(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("Patch Listing Bad ID", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/listings/invalid", bytes.NewBufferString("{}"))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "invalid")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.PatchListing(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Patch Listing Ownership Forbidden", func(t *testing.T) {
		reqBody := `{"title":"Forbidden Watch"}`
		req := httptest.NewRequest("PATCH", "/listings/1", bytes.NewBufferString(reqBody))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 2)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.PatchListing(recorder, req)

		if recorder.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", recorder.Code)
		}
	})

	t.Run("Patch Listing Not Found", func(t *testing.T) {
		reqBody := `{"title":"Patched Watch"}`
		req := httptest.NewRequest("PATCH", "/listings/999", bytes.NewBufferString(reqBody))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "999")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.PatchListing(recorder, req)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})

	t.Run("Notify Interested", func(t *testing.T) {
		store.CreateBookmark(context.Background(), 1, 1)

		req := httptest.NewRequest("POST", "/listings/1/notify-interested", nil)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.NotifyInterested(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("Notify Interested Bad ID", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/listings/invalid/notify-interested", nil)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "invalid")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.NotifyInterested(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Price Advisor", func(t *testing.T) {
		reqBody := `{"description":"Mechanical Keyboard","category":"Electronics","current_asking_price":150.00}`
		req := httptest.NewRequest("POST", "/api/price-advisor", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.PriceAdvisor(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Price Advisor Bad Input", func(t *testing.T) {
		reqBody := `{"description":"","category":"","current_asking_price":-10.0}`
		req := httptest.NewRequest("POST", "/api/price-advisor", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.PriceAdvisor(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Price Advisor Gemini API Error", func(t *testing.T) {
		oldKey := os.Getenv("GEMINI_API_KEY")
		os.Setenv("GEMINI_API_KEY", "dummy-key")
		defer os.Setenv("GEMINI_API_KEY", oldKey)

		reqBody := `{"description":"Mechanical Keyboard","category":"Electronics","current_asking_price":150.00}`
		req := httptest.NewRequest("POST", "/api/price-advisor", bytes.NewBufferString(reqBody))
		recorder := httptest.NewRecorder()

		handler.PriceAdvisor(recorder, req)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d. Body: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("Price Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings/price-check?item=Bicycle", nil)
		recorder := httptest.NewRecorder()

		handler.PriceCheck(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Price Check Missing Item", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings/price-check", nil)
		recorder := httptest.NewRecorder()

		handler.PriceCheck(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Search Listings", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/search?q=watch&category=Watches&min_price=100&max_price=500", nil)
		recorder := httptest.NewRecorder()

		handler.SearchListings(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Recommended Listings", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/listings/recommended", nil)
		recorder := httptest.NewRecorder()

		handler.GetRecommendedListings(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Delete Listing Bad ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/listings/invalid", nil)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "invalid")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteListing(recorder, req)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("Delete Listing Ownership Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/listings/1", nil)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 2)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteListing(recorder, req)

		if recorder.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", recorder.Code)
		}
	})

	t.Run("Delete Listing Not Found", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/listings/999", nil)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "999")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteListing(recorder, req)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})

	t.Run("Delete Listing", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/listings/1", nil)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", "1")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteListing(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("Logout", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/logout", nil)
		recorder := httptest.NewRecorder()

		handler.Logout(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})
}
