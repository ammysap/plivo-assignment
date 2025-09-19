package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ammysap/plivo-pub-sub/libraries/auth"
	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/ammysap/plivo-pub-sub/pubsub"
	"github.com/ammysap/plivo-pub-sub/services/gateway/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.WithContext(ctx)

	logger.Info("Starting PubSub Gateway Service...")

	// Initialize auth
	auth.InitAuth(auth.AuthTypeHMAC)

	// Initialize PubSub service (singleton)
	logger.Info("Initializing PubSub service...")
	pubsubService := pubsub.InitService(pubsub.DefaultConfig())

	// Start the service
	logger.Info("Starting PubSub service...")
	err := pubsubService.Start(ctx)
	if err != nil {
		logger.Errorw("Failed to start PubSub service", "error", err)
		log.Fatalf("cannot start pubsub service: %v", err)
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	serverDone := make(chan error, 1)
	go func() {
		logger.Info("Starting HTTP server...")
		err := app.RegisterRoutes(ctx, nil)
		serverDone <- err
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverDone:
		if err != nil {
			logger.Errorw("HTTP server error", "error", err)
			log.Fatalf("HTTP server failed: %v", err)
		}
	case sig := <-shutdown:
		logger.Infow("Received shutdown signal", "signal", sig)
	}

	// Graceful shutdown
	logger.Info("Starting graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop PubSub service
	logger.Info("Stopping PubSub service...")
	if err := pubsubService.Stop(shutdownCtx); err != nil {
		logger.Errorw("Error stopping PubSub service", "error", err)
	} else {
		logger.Info("PubSub service stopped successfully")
	}

	logger.Info("Graceful shutdown completed")
}
