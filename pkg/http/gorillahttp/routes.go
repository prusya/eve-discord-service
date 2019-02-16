package gorillahttp

import (
	"net/http"
)

// Routes adds routes and handlers to the router.
func (s *Service) Routes() {
	s.router.NotFoundHandler = http.HandlerFunc(NotFoundH)
	s.router.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedH)

	jsonAPI := s.router.PathPrefix("/api").Subrouter()
	jsonAPI.HandleFunc("/healthcheck", HealthCheckH)

	// discord service routes.
	dv1 := jsonAPI.PathPrefix("/discord/v1").Subrouter()
	dv1.HandleFunc("/auth", s.DiscordAuthH)
	dv1.HandleFunc("/auth/callback", s.DiscordAuthCallbackH)
}
