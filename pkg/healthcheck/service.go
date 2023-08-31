package healthcheck

import "context"

// Service is the interface that provides healthcheck methods
type Service interface {
	// Check repository aliveness
	Alive(ctx context.Context) error
}

type service struct {
	healthcheck HealthcheckRepository
}

// NewService creates a tranfer service with necessary dependencies
func NewService(
	healthcheck HealthcheckRepository,
) Service {
	return &service{
		healthcheck: healthcheck,
	}
}

func (s *service) Alive(ctx context.Context) error {
	if err := s.healthcheck.Ping(ctx); err != nil {
		return err
	}
	return nil
}
