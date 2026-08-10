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
	"github.com/shopspring/decimal"
)

const defaultTaskLeaseMaxSeconds = 300

type Config struct {
	DatabaseURL             string
	HTTPPort                int
	TaskLeaseMaxSeconds     int
	DealManagerAPIBaseURL   string
	OutboxPublishInterval   time.Duration
	MinimumProfit           decimal.Decimal
	MaximumBuyPrice         decimal.Decimal
	MaximumSummaryAgeSecs   int
	MaximumConcurrentTrades int
}

var (
	ErrDatabaseConfig   = errors.New("database connection url not set or incorrect")
	ErrPort             = errors.New("http server port not specified or incorrect")
	ErrTaskLeaseMax     = errors.New("task lease max seconds incorrect")
	ErrDealManagerAPI   = errors.New("dealmanager api base url not set or incorrect")
	ErrOutboxInterval   = errors.New("outbox publish interval not specified or incorrect")
	ErrMinimumProfit    = errors.New("minimum profit not specified or incorrect")
	ErrMaximumBuyPrice  = errors.New("maximum buy price not specified or incorrect")
	ErrSummaryAge       = errors.New("maximum summary age not specified or incorrect")
	ErrConcurrentTrades = errors.New("maximum concurrent trades not specified or incorrect")
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
	minimumProfit, err := decimal.NewFromString(os.Getenv("TRADE_MIN_PROFIT"))
	if err != nil || minimumProfit.IsNegative() {
		return nil, ErrMinimumProfit
	}
	maximumBuyPrice, err := decimal.NewFromString(os.Getenv("TRADE_MAX_BUY_PRICE"))
	if err != nil || maximumBuyPrice.LessThanOrEqual(decimal.Zero) {
		return nil, ErrMaximumBuyPrice
	}
	maximumSummaryAgeSecs, err := strconv.Atoi(os.Getenv("TRADE_MAX_SUMMARY_AGE_SECONDS"))
	if err != nil || maximumSummaryAgeSecs < 1 {
		return nil, ErrSummaryAge
	}
	maximumConcurrentTrades, err := strconv.Atoi(os.Getenv("TRADE_MAX_CONCURRENT_TRADES"))
	if err != nil || maximumConcurrentTrades < 1 {
		return nil, ErrConcurrentTrades
	}

	return &Config{
		DatabaseURL:             databaseURL,
		HTTPPort:                httpPort,
		TaskLeaseMaxSeconds:     taskLeaseMaxSeconds,
		DealManagerAPIBaseURL:   dealManagerAPIBaseURL,
		OutboxPublishInterval:   outboxPublishInterval,
		MinimumProfit:           minimumProfit,
		MaximumBuyPrice:         maximumBuyPrice,
		MaximumSummaryAgeSecs:   maximumSummaryAgeSecs,
		MaximumConcurrentTrades: maximumConcurrentTrades,
	}, nil
}

func apiBaseURL(value string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ErrDealManagerAPI
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}
