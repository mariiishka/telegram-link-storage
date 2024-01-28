package main

import (
	"flag"
	"log"

	tgClient "github.com/mariiishka/telegram-link-storage/clients/telegram"
	event_consumer "github.com/mariiishka/telegram-link-storage/consumer/event-consumer"
	"github.com/mariiishka/telegram-link-storage/events/telegram"
	"github.com/mariiishka/telegram-link-storage/storage/files"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 100
)

func main() {
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		files.New(storagePath),
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}

func mustToken() string {
	token := flag.String("tg-bot-token", "", "token for access to telegram bot")

	flag.Parse()
	if *token == "" {
		log.Fatal("token is not specified")
	}

	return *token
}
