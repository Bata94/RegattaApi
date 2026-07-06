package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bata94/RegattaApi/internal/config"
	DB "github.com/bata94/RegattaApi/internal/db"
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
	defer DB.ShutdownConnection()

	utils.InitEmail()
	os.MkdirAll(config.C.Paths.PublicDir, os.ModePerm)

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

	addr := config.C.Server.Host + ":" + config.C.Server.Port
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.GetRouter()))
}
