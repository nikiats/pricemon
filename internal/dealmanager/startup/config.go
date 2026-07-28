package startup

import (
	"errors"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrDatabaseConfig = errors.New("dealmanager database connection url not set")
	ErrPort           = errors.New("dealmanager http port not specified or incorrect")
	ErrEventInterval  = errors.New("dealmanager event process interval not specified or incorrect")
	ErrGoPriceMonAPI  = errors.New("gopricemon api base url not set or incorrect")
)

type Config struct {
	DatabaseURL          string
	HTTPPort             int
	EventProcessInterval time.Duration
	GoPriceMonAPIBaseURL string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err == nil {
		log.Println(".env file loaded")
	}

	databaseURL := os.Getenv("DEALMANAGER_DATABASE_URL")
	if databaseURL == "" {
		return nil, ErrDatabaseConfig
	}

	httpPort, err := strconv.Atoi(os.Getenv("DEALMANAGER_HTTP_PORT"))
	if err != nil || httpPort < 1 || httpPort > 65535 {
		return nil, ErrPort
	}

	eventProcessInterval, err := time.ParseDuration(os.Getenv("DEALMANAGER_EVENT_PROCESS_INTERVAL"))
	if err != nil || eventProcessInterval <= 0 {
		return nil, ErrEventInterval
	}

	goPriceMonAPIBaseURL, err := apiBaseURL(os.Getenv("GOPRICEMON_API_BASE_URL"))
	if err != nil {
		return nil, ErrGoPriceMonAPI
	}
	return &Config{
		DatabaseURL:          databaseURL,
		HTTPPort:             httpPort,
		EventProcessInterval: eventProcessInterval,
		GoPriceMonAPIBaseURL: goPriceMonAPIBaseURL,
	}, nil
}

func apiBaseURL(value string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ErrGoPriceMonAPI
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}
