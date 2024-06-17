package configs

// OpenTelemetry describe common settings for open telemetry
type OpenTelemetry struct {
	Enable bool `env:"OPEN_TELEMETRY_ENABLE"`
}
