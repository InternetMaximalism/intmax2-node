package configs

// GRPC describe common settings for gPRC
type GRPC struct {
	Host string `env:"GRPC_HOST" envDefault:"0.0.0.0"`
	Port string `env:"GRPC_PORT" envDefault:"10000"`
}

func (grpc *GRPC) Addr() string {
	return grpc.Host + hostPortDelimiter + grpc.Port
}
