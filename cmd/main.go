package main

import (
	"context"
	"net/http"
	"os"
	"time"

	observability "github.com/Gen-Do/lib-observability"
	platform "github.com/Gen-Do/lib-platform"
	"github.com/Gen-Do/lib-transport/listener"
	"github.com/Gen-Do/skeleton-service/internal/workers/example"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

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
	obs.SetupHTTP(r)

	// Ваши обработчики
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info(r.Context(), "Hello World!")
		w.Write([]byte("Hello World!"))
	})

	lis := listener.New(
		listener.WithIdleTimeout(10*time.Second),
		listener.WithReadTimeout(10*time.Second),
		listener.WithWriteTimeout(10*time.Second),
		listener.WithMW(middleware.RequestID),
		listener.WithLogger(log),
	)

	err := platform.Run(ctx,
		platform.WithListener(lis),
		platform.WithMux(r),
		platform.WithLogger(log),
		platform.WithEnableSignalHandling(true),
		platform.WithObservability(platform.ObservabilitySettings{
			Logger:  log,
			Metrics: nil,
		}),
		platform.WithWorkers(example.NewWorker(log)),
	)
	if err != nil {
		log.Error(log.WithError(ctx, err), "Application exited with error")
		return platform.ExitCodeFailure
	}

	log.Info(ctx, "Service stopped gracefully")

	return platform.ExitCodeSuccess
}
