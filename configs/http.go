package configs

// HTTP describe common settings for http
type HTTP struct {
	CORS []string `env:"HTTP_CORS" envSeparator:";" envDefault:"*"`
	// value in seconds
	CORSAllowAll         bool     `env:"HTTP_CORS_ALLOW_ALL" envDefault:"false"`
	CORSMaxAge           int      `env:"HTTP_CORS_MAX_AGE" envDefault:"600"`
	CORSStatusCode       int      `env:"HTTP_CORS_STATUS_CODE" envDefault:"204"`
	CORSAllowCredentials bool     `env:"HTTP_CORS_ALLOW_CREDENTIALS" envDefault:"true"`
	CORSAllowMethods     []string `env:"HTTP_CORS_ALLOW_METHODS" envSeparator:";" envDefault:"OPTIONS"`
	CORSAllowHeaders     []string `env:"HTTP_CORS_ALLOW_HEADS" envSeparator:";" envDefault:"Accept;Authorization;Content-Type;X-CSRF-Token;X-User-Id;X-Api-Key"` //nolint:lll
	CORSExposeHeaders    []string `env:"HTTP_CORS_EXPOSE_HEADS" envSeparator:";" envDefault:""`
	Host                 string   `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port                 string   `env:"HTTP_PORT" envDefault:"80"`
	TLSUse               bool     `env:"HTTP_TLS_USE" envDefault:"false"`

	CookieSecure             bool   `env:"COOKIE_SECURE"`
	CookieDomain             string `env:"COOKIE_DOMAIN" envDefault:""`
	CookieSameSiteStrictMode bool   `env:"COOKIE_SAME_SITE_STRICT_MODE"`

	CookieForAuthUse bool `env:"COOKIE_FOR_AUTH_USE"`

	// value equal CORSMaxAge
	Timeout int
}

func (http *HTTP) Addr() string {
	return http.Host + hostPortDelimiter + http.Port
}
