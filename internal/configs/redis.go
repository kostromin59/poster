package configs

type Redis struct {
	Conn     string `envconfig:"REDIS_CONN" required:"true"`
	Password string `envconfig:"REDIS_PASSWORD" required:"true"`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}
