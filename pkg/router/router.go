package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Route struct {
	Pattern     string
	Method      string
	Handler     func(w http.ResponseWriter, r *http.Request)
	Description string
}

type Router struct {
	Router   *chi.Mux
	Shutdown chan struct{}
}

func New() *Router {
	return &Router{
		Router:   chi.NewRouter(),
		Shutdown: make(chan struct{}),
	}
}

func (r *Router) WithRoutes(routes []Route) *Router {
	for _, route := range routes {
		r.Router.Method(route.Method, route.Pattern, http.HandlerFunc(route.Handler))
	}
	return r
}
