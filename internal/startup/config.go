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

const defaultTaskLeaseMaxSeconds = 300

type Config struct {
	DatabaseURL           string
	HTTPPort              int
	TaskLeaseMaxSeconds   int
	DealManagerAPIBaseURL string
	OutboxPublishInterval time.Duration
}

var (
	ErrDatabaseConfig = errors.New("database connection url not set or incorrect")
	ErrPort           = errors.New("http server port not specified or incorrect")
	ErrTaskLeaseMax   = errors.New("task lease max seconds incorrect")
	ErrDealManagerAPI = errors.New("dealmanager api base url not set or incorrect")
	ErrOutboxInterval = errors.New("outbox publish interval not specified or incorrect")
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

	taskLeaseMaxSeconds := defaultTaskLeaseMaxSeconds
	if taskLeaseMaxRaw := os.Getenv("TASK_LEASE_MAX_SECONDS"); taskLeaseMaxRaw != "" {
		taskLeaseMaxSeconds, err = strconv.Atoi(taskLeaseMaxRaw)
		if err != nil || taskLeaseMaxSeconds < 1 {
			return nil, ErrTaskLeaseMax
		}
	}

	dealManagerAPIBaseURL, err := apiBaseURL(os.Getenv("DEALMANAGER_API_BASE_URL"))
	if err != nil {
		return nil, ErrDealManagerAPI
	}
	outboxPublishInterval, err := time.ParseDuration(os.Getenv("OUTBOX_PUBLISH_INTERVAL"))
	if err != nil || outboxPublishInterval <= 0 {
		return nil, ErrOutboxInterval
	}

	return &Config{
		DatabaseURL:           databaseURL,
		HTTPPort:              httpPort,
		TaskLeaseMaxSeconds:   taskLeaseMaxSeconds,
		DealManagerAPIBaseURL: dealManagerAPIBaseURL,
		OutboxPublishInterval: outboxPublishInterval,
	}, nil
}

func apiBaseURL(value string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ErrDealManagerAPI
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}
