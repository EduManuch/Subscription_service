// @title Subscription Service API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /
package main

import (
	"net/http"
	"os"
	_ "sub_service/docs"
	"sub_service/internal/config"
	"sub_service/internal/handlers"
	"sub_service/internal/logger"
	"sub_service/internal/middleware"
	"sub_service/internal/repository"
	"sub_service/internal/service"
	"sub_service/internal/storage"

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
	defer db.Close()

	repo := repository.NewSubscriptionRepository(db)
	subscriptionService := service.NewSubscriptionService(repo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, log)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscriptions", subscriptionHandler.Create)
	mux.HandleFunc("GET /subscriptions/{id}", subscriptionHandler.GetByID)
	mux.HandleFunc("GET /subscriptions", subscriptionHandler.List)
	mux.HandleFunc("PUT /subscriptions/{id}", subscriptionHandler.Update)
	mux.HandleFunc("DELETE /subscriptions/{id}", subscriptionHandler.Delete)
	mux.HandleFunc("GET /subscriptions/total", subscriptionHandler.CalculateTotalPrice)
	mux.Handle("GET /swagger/", swagger.WrapHandler)

	muxWithMiddleware := middleware.LoggingMiddleware(log)(mux)

	log.Info("Server started", "address", cfg.AppPort)
	if err := http.ListenAndServe(":"+cfg.AppPort, muxWithMiddleware); err != nil {
		log.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
