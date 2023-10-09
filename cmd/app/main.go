package main

import (
	"log"

	"github.com/lintangbs/chat-be/config"
	"github.com/lintangbs/chat-be/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
