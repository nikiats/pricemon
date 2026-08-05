package startup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"gopricemon/internal/repository/postgres"
	"gopricemon/internal/service"
	transport "gopricemon/internal/transport/http"
)

func Run(config Config) error {
	repository, err := postgres.NewRepository(config.DatabaseURL)
	if err != nil {
		return err
	}
	defer repository.Close()

	application := service.NewService(repository, config.TaskLeaseMaxSeconds)
	if err = application.InitializeTradeSettings(config.MinimumProfit, config.MaximumSummaryAgeSecs, config.MaximumConcurrentTrades); err != nil {
		return err
	}

	// gopricemon API
	router := gin.Default()
	transport.NewHandler(
		application,
	).Register(router)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
		Handler: router,
	}

	// graceful stop
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// запуск воркера отправки событий
	publisher := service.NewOutboxPublisher(repository, config.DealManagerAPIBaseURL, config.OutboxPublishInterval)
	go publisher.Run(ctx)

	// запуск HTTP API
	errorsCh := make(chan error, 1)
	go func() {
		errorsCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errorsCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		// завершение работы по сигналу SIGTERM
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	}
}
