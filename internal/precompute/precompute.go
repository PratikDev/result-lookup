package precompute

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	precomputeUtils "github.com/pratikdev/result-lookup/internal/precompute/utils"
	redisClient "github.com/pratikdev/result-lookup/internal/redis"
	"github.com/redis/go-redis/v9"
)

func RunPrecompute(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger, examYear int) error {
	err := precomputeUtils.BatchRedisSeeding(ctx, pool, rdb, logger, examYear)
	if err != nil {
		logger.Error("batch seeding redis from postgress process failed", "error", err)
		return err
	}

	/* Count Verification */
	err = precomputeUtils.CountVerification(ctx, pool, rdb, examYear)
	if err != nil {
		logger.Error("results count verification between redis and postgress process failed", "error", err)
		return err
	}

	/* Flip publish gate */
	err = redisClient.Set(rdb, logger, "result:published", "1")
	if err != nil {
		logger.Error("publish gate flip process failed", "error", err)
		return err
	}

	logger.Info("publish gate flipped. results are now live.", "examYear", examYear)
	return nil
}