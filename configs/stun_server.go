package configs

type StunServer struct {
	NetworkType []string `env:"STUN_SERVER_NETWORK_TYPE,required" envDefault:"udp6;udp4" envSeparator:";"`
	List        []string `env:"STUN_SERVER_LIST,required" envDefault:"stun.l.google.com:19302" envSeparator:";"`
}
