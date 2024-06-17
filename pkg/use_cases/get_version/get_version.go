package get_version

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	getVersion "intmax2-node/internal/use_cases/get_version"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// uc describes use case
type uc struct {
	Version   string
	BuildTime string
}

func New(version, buildTime string) getVersion.UseCaseGetVersion {
	return &uc{
		Version:   version,
		BuildTime: buildTime,
	}
}

func (u *uc) Do(ctx context.Context) *getVersion.Version {
	const (
		hName     = "UseCase GetVersion"
		version   = "version"
		buildTime = "build_time"
	)

	_, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(version, u.Version),
			attribute.String(buildTime, u.BuildTime),
		))
	defer span.End()

	return &getVersion.Version{
		Version:   u.Version,
		BuildTime: u.BuildTime,
	}
}
