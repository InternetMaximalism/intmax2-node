package configs

import "time"

type DepositSynchronizer struct {
	TimeoutForEventWatcher time.Duration `env:"DEPOSIT_SYNCHRONIZER_TIMEOUT_FOR_EVENT_WATCHER" envDefault:"5m"`
	Enable                 bool          `env:"DEPOSIT_SYNCHRONIZER_ENABLE" envDefault:"false"`
}
