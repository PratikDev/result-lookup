package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(logger *slog.Logger) *pgxpool.Pool {
	DATABASE_URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	
	dbpool, err := pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		logger.Error("Unable to create connection pool", "error", err)
		os.Exit(1)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		logger.Error("Unable to connect to the db", "error", err)
		os.Exit(1)
	}

	logger.Info("Succesfully connected to the db")
	return dbpool
}

func InitFallbackDB(logger *slog.Logger) *pgxpool.Pool {
	DATABASE_URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	config, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		logger.Error("Unable setup fallback db config", "error", err)
		return nil
	}
	maxConn, err := strconv.Atoi(os.Getenv("PG_FALLBACK_MAX_CONN"))
	if err != nil {
		maxConn = 20
	}
	config.MaxConns = int32(maxConn)
	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Error("unable to create fallback db connection pool","error", err)
		return nil
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		logger.Error("Unable to connect to the fallback db", "error", err)
		return nil
	}

	logger.Info("Succesfully connected to the fallback db")
	return dbpool
}