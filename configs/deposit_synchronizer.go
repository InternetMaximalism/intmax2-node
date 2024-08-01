package configs

import "time"

type DepositSynchronizer struct {
	ID                     string        `env:"DEPOSIT_SYNCHRONIZER_ID"`
	Path                   string        `env:"DEPOSIT_SYNCHRONIZER_PATH" envDefault:"/app/deposit_synchronizer"`
	TimeoutForEventWatcher time.Duration `env:"DEPOSIT_SYNCHRONIZER_TIMEOUT_FOR_EVENT_WATCHER" envDefault:"5m"`
}
