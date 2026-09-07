package config

import (
	"pricemon/internal/config"
)

type Config struct {
	Server  config.Server
	DB      config.DB
	Service config.Service
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := config.Load(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
