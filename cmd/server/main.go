package main

import (
	"context"
	"log/slog"
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
	setupLogger()

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
			slog.Error("Error shutting down DB connection", "err", err)
		}
	}()

	utils.InitEmail()
	if err := os.MkdirAll(config.C.Paths.PublicDir, os.ModePerm); err != nil {
		slog.Error("Error creating public dir", "err", err)
	}

	go mailer.RunWorker(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			slog.Info("Received signal", "sig", sig)
			if sig == os.Interrupt {
				if err := DB.ShutdownConnection(); err != nil {
					slog.Error("Error shutting down DB connection", "err", err)
				}
				os.Exit(0)
			}
		}
	}()

	addr := config.C.Server.Host + ":" + config.C.Server.Port
	slog.Info("Starting server", "addr", addr)
	if err := http.ListenAndServe(addr, server.GetRouter()); err != nil {
		slog.Error("Server failed", "err", err)
		os.Exit(1)
	}
}

func setupLogger() {
	level := slog.LevelWarn
	if config.C.Env == "dev" {
		level = slog.LevelDebug
	}

	if config.C.Log.Level != "" {
		var parsed slog.Level
		if err := parsed.UnmarshalText([]byte(config.C.Log.Level)); err != nil {
			slog.Warn("Invalid LOG_LEVEL, falling back to default", "value", config.C.Log.Level, "err", err)
		} else {
			level = parsed
		}
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
	slog.Info("Logger initialized", "level", level, "env", config.C.Env)
}
