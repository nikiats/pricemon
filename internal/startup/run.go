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

	router := gin.Default()
	transport.NewHandler(service.NewService(repository), config.AdminPassword).Register(router)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
		Handler: router,
	}

	errorsCh := make(chan error, 1)
	go func() {
		errorsCh <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errorsCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	}
}
