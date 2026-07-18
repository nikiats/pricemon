package startup

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	HTTPPort      int
	AdminPassword string
}

var (
	ErrDatabaseConfig = errors.New("database connection url not set or incorrect")
	ErrPort           = errors.New("http server port not specified or incorrect")
	ErrAdminPassword  = errors.New("admin password not set")
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

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		return nil, ErrAdminPassword
	}

	return &Config{
		DatabaseURL:   databaseURL,
		HTTPPort:      httpPort,
		AdminPassword: adminPassword,
	}, nil
}
