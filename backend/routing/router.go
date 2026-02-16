package routing

import (
	routes_base "aura/routing/base"
	"aura/routing/middleware"
	"sync"

	"github.com/go-chi/chi/v5"
)

// OnboardingComplete can be set by main to perform:
// - validation preflight
// - DB init
// - cron start
// - router swap
var OnboardingComplete func()

// ensure finalize only runs once
var onboardingFinalizeOnce sync.Once

func NewRouter() *chi.Mux {
	// Create a new router
	r := chi.NewRouter()

	// Configure the router with middlewares
	middleware.Configure(r)

	// Add the routes to the router
	AddRoutes(r)

	// If the route is not found, return a JSON response
	r.NotFound(routes_base.NotFound)
	r.MethodNotAllowed(routes_base.MethodNotAllowed)

	return r
}
