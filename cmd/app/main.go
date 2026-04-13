package main

import (
	"log"
	"sub_service/internal/config"
	"sub_service/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := storage.NewPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("App started successfully")
}
