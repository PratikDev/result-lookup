package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/pratikdev/result-lookup/internal/database"
	"github.com/pratikdev/result-lookup/internal/precompute"
	redisClient "github.com/pratikdev/result-lookup/internal/redis"
)

const MIN_EXAM_YEAR = 2000
const MAX_EXAM_YEAR = 2200

func main() {
	examYear := flag.Int("year", 0, "exam year to precompute")
	flag.Parse()

	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)

	if (*examYear < MIN_EXAM_YEAR) || (*examYear > MAX_EXAM_YEAR) {
		logger.Error("invalid exam year. use -year flag (must be between 2000 to 2200)")
		os.Exit(1)
	}

	pool := database.InitDB(logger)
	rdb := redisClient.InitRedis(logger)
	ctx := context.Background()

	defer pool.Close()
	defer rdb.Close()

	err := precompute.RunPrecompute(ctx, pool, rdb, logger, *examYear)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}