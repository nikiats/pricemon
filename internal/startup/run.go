package startup

import (
	"fmt"

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

	return router.Run(fmt.Sprintf(":%d", config.HTTPPort))
}
