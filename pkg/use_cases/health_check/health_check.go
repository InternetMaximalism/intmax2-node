package health_check

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	healthCheck "intmax2-node/internal/use_cases/health_check"

	"github.com/dimiro1/health"
)

// uc describes use case
type uc struct {
	hc *health.Handler
}

func New(hc *health.Handler) healthCheck.UseCaseHealthCheck {
	return &uc{
		hc: hc,
	}
}

func (u *uc) Do(ctx context.Context) *healthCheck.HealthCheck {
	const (
		hName = "UseCase HealthCheck"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	return &healthCheck.HealthCheck{Success: u.hc.Check(spanCtx).IsUp()}
}
