package configs

import "time"

type BlockValidityProver struct {
	TimeoutForEventWatcher time.Duration `env:"BLOCK_VALIDITY_PROVER_EVENT_WATCHER_LIFETIME" envDefault:"1m"`
}
