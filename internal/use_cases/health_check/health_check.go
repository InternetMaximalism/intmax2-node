package health_check

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_health_check.go -package=mocks -source=health_check.go

type HealthCheck struct {
	Success bool
}

// UseCaseHealthCheck describes HealthCheck contract.
type UseCaseHealthCheck interface {
	Do(ctx context.Context) *HealthCheck
}
