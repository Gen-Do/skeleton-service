package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	ctx = log.WithField(ctx, "port", port)
	log.Info(ctx, "Server starting")
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)

	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.Canceled) {
		log.Error(log.WithError(ctx, ctx.Err()), "Application stopped with error")
		return fail
	}

	log.Info(ctx, "Service stopped")

	return success
}
