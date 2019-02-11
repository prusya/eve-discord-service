package http

// Service defines an interface of how to ineract with http service.
type Service interface {
	Start()
	Stop()
}
