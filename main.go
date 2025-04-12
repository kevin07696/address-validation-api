package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"address-validator/adapters"
	"address-validator/config"
	"address-validator/handlers"
	"address-validator/services"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting address validator service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create Google Maps adapter
	mapsAdapter, err := adapters.NewGoogleMapsAdapter(cfg, logger)
	if err != nil {
		logger.Error("failed to create Google Maps adapter", "error", err)
		os.Exit(1)
	}

	// Create address service
	addressService := services.NewAddressService(mapsAdapter, logger, cfg)

	// Create address handler
	addressHandler := handlers.NewAddressHandler(addressService, cfg, logger)

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", addressHandler.ValidateAddress)

	// Add basic health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting HTTP server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server exited properly")
}
