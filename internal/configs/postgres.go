package configs

import "fmt"

type Postgres struct {
	Host     string `envconfig:"POSTGRES_HOST" required:"true"`
	Port     string `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string `envconfig:"POSTGRES_USER" required:"true"`
	Password string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	DB       string `envconfig:"POSTGRES_DB" required:"true"`
}

func (p Postgres) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s database=%s",
		p.Host, p.Port, p.User, p.Password, p.DB,
	)
}
