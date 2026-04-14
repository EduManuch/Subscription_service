package main

import (
	"log"
	"net/http"
	"sub_service/internal/config"
	"sub_service/internal/handlers"
	"sub_service/internal/repository"
	"sub_service/internal/service"
	"sub_service/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := storage.NewPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewSubscriptionRepository(db)
	subscriptionService := service.NewSubscriptionService(repo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscriptions", subscriptionHandler.Create)

	log.Printf("Server started on: %s", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
