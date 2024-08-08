package configs

import "time"

type BlockPostService struct {
	TimeoutForEventWatcher time.Duration `env:"BLOCK_POST_SERVICE_EVENT_WATCHER_LIFETIME" envDefault:"1s"`
}
