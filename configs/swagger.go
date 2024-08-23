package configs

import (
	"regexp"
)

const (
	SwaggerBlockBuilderPath        = "block_builder/apidocs.swagger.json"
	SwaggerOpenAPIBlockBuilderPath = "OpenAPI/block_builder_service"

	SwaggerStoreVaultPath        = "store_vault/apidocs.swagger.json"
	SwaggerOpenAPIStoreVaultPath = "OpenAPI/store_vault_service"

	SwaggerWithdrawalPath        = "withdrawal/apidocs.swagger.json"
	SwaggerOpenAPIWithdrawalPath = "OpenAPI/withdrawal_service"

	swaggerBuildVersion = `SWAGGER_VERSION`
	swaggerHostURL      = `SWAGGER_HOST_URL`
	swaggerBasePATH     = `SWAGGER_BASE_PATH`
)

type Swagger struct {
	HostURL  string `env:"SWAGGER_HOST_URL" envDefault:"127.0.0.1:8780"`
	BasePath string `env:"SWAGGER_BASE_PATH" envDefault:"/"`

	RegexpBuildVersion *regexp.Regexp
	RegexpHostURL      *regexp.Regexp
	RegexpBasePATH     *regexp.Regexp
}

func (s *Swagger) Prepare() {
	s.RegexpBuildVersion = regexp.MustCompile(swaggerBuildVersion)
	s.RegexpHostURL = regexp.MustCompile(swaggerHostURL)
	s.RegexpBasePATH = regexp.MustCompile(swaggerBasePATH)
}
