package configs

type LOG struct {
	Level      string `env:"LOG_LEVEL" envDefault:"debug"`
	JSON       bool   `env:"LOG_JSON" envDefault:"false"`
	TimeFormat string `env:"LOG_TIME_FORMAT" envDefault:"2006-01-02T15:04:05Z"`
	IsLogLine  bool   `env:"IS_LOG_LINES"`
}
