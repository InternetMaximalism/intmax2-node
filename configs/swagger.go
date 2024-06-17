package configs

import (
	"regexp"
)

const (
	SwaggerPath        = "node/apidocs.swagger.json"
	SwaggerOpenAPIPath = "OpenAPI"

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
