package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Server struct {
	Port string `env:"PORT" envDefault:"8080"`
}

type DB struct {
	URL string `env:"DATABASE_URL,required"`
}

func Load[T any](cfg *T) error {
	_ = godotenv.Load()
	return env.Parse(cfg)
}
