package configs

import "time"

type Worker struct {
	ID                                 string        `env:"WORKER_ID"`
	Path                               string        `env:"WORKER_PATH" envDefault:"/app/worker"`
	MaxCounter                         int32         `env:"WORKER_MAX_COUNTER" envDefault:"20"`
	PathCleanInStart                   bool          `env:"WORKER_PATH_CLEAN_IN_START"`
	CurrentFileLifetime                time.Duration `env:"WORKER_CURRENT_FILE_LIFETIME" envDefault:"1m"`
	TimeoutForCheckCurrentFile         time.Duration `env:"WORKER_TIMEOUT_FOR_CHECK_CURRENT_FILE" envDefault:"10s"`
	TimeoutForSignaturesAvailableFiles time.Duration `env:"WORKER_TIMEOUT_FOR_SIGNATURES_AVAILABLE_FILES" envDefault:"15s"`
	MaxCounterOfUsers                  int           `env:"WORKER_MAX_COUNTER_OF_USERS" envDefault:"128"`
}
