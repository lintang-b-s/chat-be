package main

import (
	"github.com/google/uuid"
	"log"

	"github.com/lintangbs/chat-be/config"
	"github.com/lintangbs/chat-be/internal/app"
)

var ServerName string

func init() {
	ServerName = InitApp()
}
func InitApp() string {
	return "chat-server" + uuid.New().String()
}

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
