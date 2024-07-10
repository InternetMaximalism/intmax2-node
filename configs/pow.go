package configs

type PoW struct {
	Difficulty uint64 `env:"POW_DIFFICULTY" envDefault:"4000"`
	Workers    int    `env:"POW_WORKERS" envDefault:"2"`
}
