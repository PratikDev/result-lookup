package main

import (
	"log/slog"
	"net/http"
	"os"

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

	defer rdb.Close() // close connection when app exists

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetHealth(w, r)
	})
	mux.HandleFunc("GET /result", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetResult(w, r, rdb, logger, validate)
	})

	http.ListenAndServe(":" + os.Getenv("PORT"), mux)
}