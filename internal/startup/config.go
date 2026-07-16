package startup

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	HTTPPort    int
}

var (
	ErrDatabaseConfig = errors.New("database connection url not set or incorrect")
	ErrPort           = errors.New("http server port not specified or incorrect")
)

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err == nil {
		log.Println(".env file loaded")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, ErrDatabaseConfig
	}

	httpPortRaw := os.Getenv("HTTP_PORT")
	if httpPortRaw == "" {
		return nil, ErrPort
	}
	httpPort, err := strconv.Atoi(httpPortRaw)
	if err != nil || httpPort < 1 || httpPort > 65535 {
		return nil, ErrPort
	}

	return &Config{
		DatabaseURL: databaseURL,
		HTTPPort:    httpPort,
	}, nil
}
