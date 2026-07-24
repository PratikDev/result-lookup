package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pratikdev/result-lookup/internal/handlers"
	redisClient "github.com/pratikdev/result-lookup/internal/redis"
)

func main() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	rdb := redisClient.InitRedis(logger)
	validate := validator.New()
	mux := http.NewServeMux()

	defer func() {
		logger.Info("Closing downstream infra connections...")
		rdb.Close()
	}()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetHealth(w, r)
	})
	mux.HandleFunc("GET /result", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetResult(w, r, rdb, logger, validate)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// server configure
	server := &http.Server{
		Addr: ":" + port,
		Handler: mux,
		ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// offload server-listening to a separate goroutine
	go func () {
		logger.Info("HTTP server spinning up", slog.String("port", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Critical server execution fault", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// register signal traps for standard os termination
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// block main thread until an active termination signal arrives
	<-ctx.Done()
	logger.Info("Shutdown signal captured. Halting inbound traffic ingestion...")

	// create a 10sec deadline context to complete in-flight requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown with active connections dangling", slog.Any("error", err))
	} else {
		logger.Info("All in-flight requests completed successfully")
	}
}