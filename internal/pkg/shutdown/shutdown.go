package shutdown

import (
	"context"
	"github.com/Gen-Do/skeleton-service/internal/pkg/server"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func GracefulShutdown(srv *server.Server, logger logrus.FieldLogger) {
	// Создание канала для получения сигналов ОС
	quit := make(chan os.Signal, 1)

	// Регистрация канала для получения определенных сигналов
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем выполнение до получения сигнала
	<-quit
	logger.Info("Shutting down server...")

	// Создание контекста с таймаутом для завершения работы
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Попытка корректного завершения работы сервера
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
		return
	}

	logger.Info("Server exited gracefully")
}
