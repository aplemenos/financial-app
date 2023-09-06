package healthchecks

import "context"

// HealthcheckRepository provides access on the store
type HealthcheckRepository interface {
	Ping(ctx context.Context) error
}
