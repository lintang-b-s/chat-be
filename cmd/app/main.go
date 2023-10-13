package main

import (
	"log"

	"github.com/lintangbs/chat-be/config"
	"github.com/lintangbs/chat-be/internal/app"
)

// @title API
// @version 2.0
// @description Chat Api.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email lintangbirdasaputra23@mail.com
// @license.name Apache 2.0
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
