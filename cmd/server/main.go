package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bata94/RegattaApi/internal/config"
	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/mailer"
	"github.com/bata94/RegattaApi/internal/server"
	"github.com/bata94/RegattaApi/internal/utils"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	config.Load()
	setupLogger()

	DB.InitConnection(DB.DBServerOptions{
		Host:               config.C.DB.Host,
		Port:               config.C.DB.Port,
		User:               config.C.DB.User,
		Password:           config.C.DB.Password,
		Name:               config.C.DB.Name,
		Sslmode:            config.C.DB.SSLMode,
		PoolMaxConns:       config.C.DB.PoolMaxConns,
		PoolMinConns:       config.C.DB.PoolMinConns,
		PoolMaxIdleSeconds: config.C.DB.PoolMaxIdleSeconds,
		ConnectTimeoutSec:  config.C.DB.ConnectTimeoutSec,
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mailer.RunWorker(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-c
		slog.Info("Received signal, shutting down", "sig", sig)
		cancel()
	}()

	mainAddr := config.C.Server.Host + ":" + config.C.Server.Port
	wsAddr := config.C.Server.Host + ":" + config.C.Server.WSPort

	mainServer := &http.Server{
		Addr:    mainAddr,
		Handler: server.GetRouter(),
	}

	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws/zeitnahme", api_v1.HandleZeitnahmeWS)
	wsServer := &http.Server{
		Addr:    wsAddr,
		Handler: wsMux,
	}

	errCh := make(chan error, 2)
	var fatalErr error

	go func() {
		slog.Info("Starting main server", "addr", mainAddr)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		slog.Info("Starting WebSocket server", "addr", wsAddr)
		if err := wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down servers...")
	case err := <-errCh:
		fatalErr = err
		slog.Error("Server exited unexpectedly, shutting down", "err", err)
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := mainServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Main server shutdown error", "err", err)
	}
	if err := wsServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("WebSocket server shutdown error", "err", err)
	}

	if err := DB.ShutdownConnection(); err != nil {
		slog.Error("Error shutting down DB connection", "err", err)
	}

	slog.Info("Shutdown complete")

	if fatalErr != nil {
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
