package configs

import "time"

type BlockPostService struct {
	TimeoutForEventWatcher time.Duration `env:"BLOCK_POST_SERVICE_EVENT_WATCHER_LIFETIME" envDefault:"1s"`
	TimeoutForPostingBlock time.Duration `env:"BLOCK_POST_SERVICE_POST_BLOCK_LIFETIME" envDefault:"3s"`
}
