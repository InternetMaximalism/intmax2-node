package configs

// APP describes app meta
type APP struct {
	PrintConfig bool `env:"PRINT_CONFIG"`

	CADomainName        string `env:"CA_DOMAIN_NAME" envDefault:"x.test.example.com"`
	PEMPathCACert       string `env:"PEM_PATH_CA_CERT" envDefault:"scripts/x509/ca_cert.pem"`
	PEMPathServCert     string `env:"PEM_PATH_SERV_CERT" envDefault:"scripts/x509/server_cert.pem"`
	PEMPathServKey      string `env:"PEM_PATH_SERV_KEY" envDefault:"scripts/x509/server_key.pem"`
	PEMPAthCACertClient string `env:"PEM_PATH_CA_CERT_CLIENT" envDefault:"scripts/x509/client_ca_cert.pem"`
	PEMPathClientCert   string `env:"PEM_PATH_CLIENT_CERT" envDefault:"scripts/x509/client_cert.pem"`
	PEMPathClientKey    string `env:"PEM_PATH_CLIENT_KEY" envDefault:"scripts/x509/client_key.pem"`
}
