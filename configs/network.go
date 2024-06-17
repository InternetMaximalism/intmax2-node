package configs

import "time"

const (
	NatDiscoverLifeTime = 3 * time.Hour
	NatDiscoverReCheck  = time.Hour
)

type Network struct {
	Domain   string `env:"NETWORK_DOMAIN"`
	Port     int    `env:"NETWORK_PORT"`
	HTTPSUse bool   `env:"NETWORK_HTTPS_USE" envDefault:"false"`
}
