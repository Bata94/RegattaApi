package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bata94/RegattaApi/internal/config"
	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/mailer"
	"github.com/bata94/RegattaApi/internal/server"
	"github.com/bata94/RegattaApi/internal/utils"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	config.Load()

	DB.InitConnection(DB.DBServerOptions{
		Host:     config.C.DB.Host,
		Port:     config.C.DB.Port,
		User:     config.C.DB.User,
		Password: config.C.DB.Password,
		Name:     config.C.DB.Name,
		Sslmode:  config.C.DB.SSLMode,
	})
	defer func() {
		if err := DB.ShutdownConnection(); err != nil {
			log.Printf("Error shutting down DB connection: %v", err)
		}
	}()

	utils.InitEmail()
	if err := os.MkdirAll(config.C.Paths.PublicDir, os.ModePerm); err != nil {
		log.Printf("Error creating public dir: %v", err)
	}

	go mailer.RunWorker(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Println("Received signal:", sig)
			if sig == os.Interrupt {
				if err := DB.ShutdownConnection(); err != nil {
					log.Printf("Error shutting down DB connection: %v", err)
				}
				os.Exit(0)
			}
		}
	}()

	addr := config.C.Server.Host + ":" + config.C.Server.Port
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.GetRouter()))
}
