package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shubhvish4495/basilisk/pkg/helper"
	"github.com/shubhvish4495/basilisk/pkg/rest"

	"github.com/gorilla/mux"
)

// main is the entry point for the application. It sets up an HTTP server
// with specific read and write timeouts, registers a shutdown function,
// and starts the server in a separate goroutine. It also listens for
// system interrupt signals (SIGINT, SIGTERM) to gracefully shut down
// the server when such a signal is received.
func main() {

	// logger initializes a new logger instance using slog with a JSON handler
	// that outputs to the standard output (os.Stdout). The handler options are
	// set to the default values.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	helper.InitLogger(logger)

	shutDownFuncWait := make(chan struct{})

	r := mux.NewRouter()

	srv := &http.Server{
		Addr:         ":4444",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	srv.RegisterOnShutdown(func() {
		logger.Info("Shutting down...")
		shutDownFuncWait <- struct{}{}
		logger.Info("Shutting down complete")
	})

	rest.RegisterRoutes(r)

	// Adding middleware to the router
	r.Use(rest.LoggingMiddleware)
	r.Use(rest.RecoveryMiddleware)

	go func() {
		logger.Info("server running on port 4444")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	<-shutDownFuncWait
	logger.Info("Server shutdown gracefully")

}
