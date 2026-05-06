package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/server"
	"github.com/bata94/RegattaApi/internal/utils"

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

	utils.InitEmail()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Println("Received signal:", sig)
			if sig == os.Interrupt {
				DB.ShutdownConnection()
				os.Exit(0)
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.GetRouter()))
}
