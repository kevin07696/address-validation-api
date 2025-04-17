package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"address-validator/adapters"
	"address-validator/config"
	"address-validator/handlers"
	"address-validator/services"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	env := config.LoadConfig()

	infraConfig := env.NewInfraConfig()

	// Initialize logger
	loggerConfig := env.NewLoggerConfig(infraConfig.Environment)

	logger, err := config.NewLogger(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to implement logger: %v", err)
	}

	logger.Info("starting address validator service")

	// Create Google Maps adapter
	mapConfig := env.NewMapConfig(logger)

	addressAdapter, err := adapters.NewGoogleAddressValidationAdapter(mapConfig, logger)
	if err != nil {
		logger.Error("failed to create Google Address Validation adapter", zap.Error(err))
		os.Exit(1)
	}

	// Create address service
	addressService := services.NewAddressService(addressAdapter, logger, mapConfig)

	// Create address handler
	rateLimitConfig := env.NewRateLimitConfig(logger)
	rateLimiter := handlers.NewRateLimiter(rateLimitConfig)
	addressHandler := handlers.NewAddressHandler(addressService, rateLimiter, infraConfig, logger)

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", addressHandler.ValidateAddress)

	// Add basic health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", infraConfig.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting HTTP server", zap.Uint16("port", infraConfig.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", zap.Error(err))
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
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited properly")
}
