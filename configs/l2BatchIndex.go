package configs

import "time"

type L2BatchIndex struct {
	L2BlockNumberTimeout time.Duration `env:"L2_BLOCK_NUMBER_TIMEOUT,required" envDefault:"10m"`
	L2BatchIndexTimeout  time.Duration `env:"L2_BATCH_INDEX_TIMEOUT,required" envDefault:"20m"`
}
