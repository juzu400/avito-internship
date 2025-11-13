package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juzu400/avito-internship/internal/config"
	"github.com/juzu400/avito-internship/internal/logger"
	"github.com/juzu400/avito-internship/internal/migrations"
	"github.com/juzu400/avito-internship/internal/repository"
	"github.com/juzu400/avito-internship/internal/service"
	httptransport "github.com/juzu400/avito-internship/internal/transport/http"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(logger.Config{Level: cfg.LogLevel})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := repository.NewPostgresDB(ctx, cfg.DBDSN)
	if err != nil {
		log.Error("failed to init db", slog.Any("err", err))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("db connected")

	// Миграции
	if err := migrations.Apply(ctx, log, db.Pool, "migrations"); err != nil {
		log.Error("failed to apply migrations", slog.Any("err", err))
		os.Exit(1)
	}
	log.Info("migrations applied")

	services := service.NewServices(log)
	router := httptransport.NewRouter(log, services)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		log.Info("server starting", slog.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", slog.Any("err", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("server shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", slog.Any("err", err))
	}
}
