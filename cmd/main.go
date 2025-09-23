package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	observability "github.com/Gen-Do/lib-observability"
	"github.com/Gen-Do/lib-observability/env"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	os.Exit(run())
}

const (
	success = 0
	fail    = 1
)

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	obs := observability.MustNew(ctx)
	defer obs.Shutdown(ctx)

	log := obs.GetLogger()
	log.Info(ctx, "Initializing service")

	// Пример использования БД
	//db, err := gorm.Open(postgres.Open(os.Getenv("DEP_DATABASE_DSN")), &gorm.Config{})
	//if err != nil {
	//	log.WithError(err).Fatal("Failed to connect to database")
	//	return fail
	//}

	// Настройка HTTP сервера
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	obs.SetupHTTP(r)

	// Ваши обработчики
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info(r.Context(), "Hello World!")
		w.Write([]byte("Hello World!"))
	})

	port := env.Get("PORT", 8080)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	// Запускаем сервер в отдельной горутине
	serverErr := make(chan error, 1)
	go func() {
		ctx = log.WithField(ctx, "port", port)
		log.Info(ctx, "Server starting")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Ждём сигнал завершения или ошибку сервера
	select {
	case err := <-serverErr:
		log.Error(log.WithError(ctx, err), "Server failed to start")
		return fail
	case <-ctx.Done():
		log.Info(ctx, "Shutdown signal received")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error(log.WithError(ctx, err), "Server shutdown failed")
		return fail
	}

	log.Info(ctx, "Service stopped gracefully")

	return success
}
