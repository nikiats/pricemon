package config

import (
	"time"

	"pricemon/internal/config"
)

type Config struct {
	Server config.Server
	DB     config.DB
	Auth   Auth
}

type Auth struct {
	JWTSecret      string        `env:"JWT_SECRET,required"`
	AccessTokenTTL time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
	BcryptCost     int           `env:"BCRYPT_COST" envDefault:"12"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := config.Load(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
