package precomputeUtils

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pratikdev/result-lookup/internal/models"
	redisClient "github.com/pratikdev/result-lookup/internal/redis"
	"github.com/redis/go-redis/v9"
)

func BatchRedisSeeding(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger, examYear int) error {
	lastID := 0
	for {
		query := `SELECT id, roll, reg, student_name, institution_name, board_name, exam_year, gpa, is_passed FROM results
		WHERE exam_year = @examYear AND id > @id
		ORDER BY id
		LIMIT 50000`
		args := pgx.NamedArgs{
			"examYear": examYear,
			"id": lastID,
		}

		rows, err := pool.Query(ctx, query, args)
		if err != nil {
			return err
		}

		// collect rows
		resultsBatch, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Result])
		if err != nil {
			return err
		}

		if len(resultsBatch) == 0 {
			break
		}

		/* SET each row in redis */
		for _, result := range resultsBatch {
			// json marshel the row
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				return err
			}
	
			resultJsonString := string(jsonBytes)
			redisKey := fmt.Sprintf(`result:%d:%d:%d`,
				result.Roll,
				result.Reg,
				examYear,
			)
			err = redisClient.Set(rdb, logger, redisKey, resultJsonString)
			if err != nil {
				return err
			}
		}

		logger.Info("batch seeded", "lastID", lastID, "batchSize", len(resultsBatch))

		lastID = resultsBatch[len(resultsBatch)-1].ID
	}

	return nil
}

func CountVerification(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, examYear int) error {
	/* Count psg results */
	query := `SELECT COUNT(*) FROM results WHERE exam_year = @examYear`
	args := pgx.NamedArgs{
		"examYear": examYear,
	}

	var psgResultsCount int
	err := pool.QueryRow(ctx, query, args).Scan(&psgResultsCount)
	if err != nil {
		return err
	}

	/* Count redis results */
	redisResultsCount := 0
	pattern := fmt.Sprintf("result:*:*:%d", examYear)
	iter := rdb.Scan(ctx, 0, pattern, 1000).Iterator()

	for iter.Next(ctx) {
		redisResultsCount++
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if psgResultsCount != redisResultsCount {
		return fmt.Errorf("verification failed: postgres and redis results count did not match")
	}
	
	return nil
}