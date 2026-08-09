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

	"gopricemon/internal/dealmanager/gopricemon"
	"gopricemon/internal/dealmanager/repository/postgres"
	app "gopricemon/internal/dealmanager/service"
	transport "gopricemon/internal/dealmanager/transport/http"
)

func Run(config Config) error {
	repository, err := postgres.NewRepository(config.DatabaseURL)
	if err != nil {
		return err
	}
	defer repository.Close()

	client := gopricemon.NewClient(config.GoPriceMonAPIBaseURL)
	inboxWorker := app.NewEventsWorker(repository, client)
	taskWorker := app.NewTaskWorker(repository, client)
	sequentialTasksService := app.NewSequentialTasksService(repository)
	router := gin.Default()
	transport.NewHandler(inboxWorker, sequentialTasksService).Register(router)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
		Handler: router,
	}

	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()

	// запуск воркера
	go inboxWorker.RunEventWorker(workerCtx, config.EventProcessInterval)
	go taskWorker.RunTaskWorker(workerCtx, config.EventProcessInterval)

	errorsCh := make(chan error, 1)
	// запуск HTTP API
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
	// сигнал завершения работы
	case <-ctx.Done():
		// остановка воркера
		stopWorker()

		// остановка HTTP API
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
