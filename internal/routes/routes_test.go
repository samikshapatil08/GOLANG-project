package routes

import (
	"testing"

	"switchboard/internal/handlers"
	"switchboard/internal/repository"
)

func TestSetup(t *testing.T) {
	store := repository.NewInMemoryStore()
	handler := handlers.NewHandler(store, store, "secret")
	router := Setup(handler)
	if router == nil {
		t.Error("expected router to not be nil")
	}
}
