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

	"github.com/shubhvish4495/basilisk/pkg/config"
	"github.com/shubhvish4495/basilisk/pkg/db"
	"github.com/shubhvish4495/basilisk/pkg/helper"
	"github.com/shubhvish4495/basilisk/pkg/rest"

	"github.com/gorilla/mux"
)

// main is the entry point of the application. It initializes the logger, configuration,
// database, and HTTP server. It also sets up signal handling for graceful shutdown.
// The server can start with or without TLS based on the presence of certificate files.
// Middleware for logging, recovery, and authentication is added to the router.
// On receiving a termination signal, the server shuts down gracefully, running
// any registered shutdown functions before exiting.
func main() {

	// shutDownFuncWait is a channel used to wait for the shutdown function to complete
	// before exiting the application.
	shutDownFuncWait := make(chan struct{})
	shutDownFuncs := make([]func() error, 0)

	// logger initializes a new logger instance using slog with a JSON handler
	// that outputs to the standard output (os.Stdout). The handler options are
	// set to the default values.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	helper.InitLogger(logger)

	// Initialize the configuration
	err := config.Load()
	if err != nil {
		logger.Error("Error initializing config", "error", err)
		os.Exit(1)
	}

	// Initialize the database and add the close function
	// on shutdown by adding it to the shutdown functions
	_, close := db.GetDb()
	shutDownFuncs = append(shutDownFuncs, close)

	r := mux.NewRouter()
	srv := &http.Server{
		Addr:         ":4444",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	srv.RegisterOnShutdown(func() {
		logger.Info("Shutting down...")
		// Run the shutdown functions
		for _, f := range shutDownFuncs {
			if err := f(); err != nil {
				logger.Error("Error shutting down", "error", err)
			}
		}
		shutDownFuncWait <- struct{}{}
		logger.Info("Shutting down complete")
	})

	rest.RegisterRoutes(r)

	// Adding middleware to the router
	r.Use(rest.LoggingMiddleware)
	r.Use(rest.RecoveryMiddleware)
	r.Use(rest.AuthMiddleware)

	// we start the server in a goroutine so that we can listen for
	// termination signals in the main goroutine. If we have provided
	// TLS certificate files, we start the server with TLS. We use
	// the ListenAndServeTLS method of the server instance. If the
	// server is started without TLS, we use the ListenAndServe method.
	// We also check if the server fails to start, and if it does, we
	// log the error and exit the application.
	go func() {
		if helper.CheckFileExist("cert.pem") && helper.CheckFileExist("key.pem") {
			logger.Info("Starting server with TLS on port 4444")
			if err := srv.ListenAndServeTLS(config.GetConfig().TlsConfig.CertFile, config.GetConfig().TlsConfig.KeyFile); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		} else {
			logger.Info("Starting server on port 4444")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	// wait for the shutdown functions to complete
	<-shutDownFuncWait
	logger.Info("Server shutdown gracefully")

}
