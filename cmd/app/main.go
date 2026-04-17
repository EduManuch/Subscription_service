// @title Subscription Service API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	_ "sub_service/docs"
	"sub_service/internal/config"
	"sub_service/internal/handlers"
	"sub_service/internal/logger"
	"sub_service/internal/middleware"
	"sub_service/internal/repository"
	"sub_service/internal/service"
	"sub_service/internal/storage"
	"syscall"
	"time"

	swagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	db, err := storage.NewPostgres(cfg)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	repo := repository.NewSubscriptionRepository(db)
	subscriptionService := service.NewSubscriptionService(repo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, log)
	router := NewRouter(subscriptionHandler, log)

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		log.Info("Server started", "address", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown server gracefully", "error", err)
	} else {
		log.Info("server stopped gracefully")
	}

	db.Close()
	log.Info("database connection closed")
}

func NewRouter(sh *handlers.SubscriptionHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscriptions", sh.Create)
	mux.HandleFunc("GET /subscriptions/{id}", sh.GetByID)
	mux.HandleFunc("GET /subscriptions", sh.List)
	mux.HandleFunc("PUT /subscriptions/{id}", sh.Update)
	mux.HandleFunc("DELETE /subscriptions/{id}", sh.Delete)
	mux.HandleFunc("GET /subscriptions/total", sh.CalculateTotalPrice)
	mux.Handle("GET /swagger/", swagger.WrapHandler)

	muxWithMiddleware := middleware.LoggingMiddleware(logger)(mux)

	return muxWithMiddleware
}
