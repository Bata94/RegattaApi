package main

import (
	"os"
	"os/signal"
	"strconv"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/server"
	"github.com/gofiber/fiber/v2/log"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	DB.InitConnection(DB.DBServerOptions{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Sslmode:  os.Getenv("DB_SSLMODE"),
	})
	defer DB.ShutdownConnection()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Error(sig)
			if sig == os.Interrupt {
				DB.ShutdownConnection()
				os.Exit(0)
			}
		}
	}()

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port <= 0 {
		port = 3000
	}
	server.Init(true, true, port)
}
