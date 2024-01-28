package main

import (
	"log"

	tgClient "github.com/mariiishka/telegram-link-storage/clients/telegram"
	event_consumer "github.com/mariiishka/telegram-link-storage/consumer/event-consumer"
	"github.com/mariiishka/telegram-link-storage/events/telegram"
	"github.com/mariiishka/telegram-link-storage/internal/config"
	"github.com/mariiishka/telegram-link-storage/internal/storage/sqlite"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "files_storage"
	batchSize   = 100
)

func main() {
	cfg := config.MustLoad()

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Fatal("failed to init storage: ", err)
	}

	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, cfg.Token),
		storage,
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}
