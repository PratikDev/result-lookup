package testutils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pratikdev/result-lookup/migrations"
)

func InitTestDB() *pgxpool.Pool {
	DATABASE_URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_TEST_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	dbpool, err := pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		log.Fatalf("Unable to create test connection pool: %v", err)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to connect to the test database: %v", err)
	}

	log.Print("Successfully connected to the test database")

	return dbpool
}

func SchemaSetup(pool *pgxpool.Pool) {
	schemaSQL, err := migrations.GetMigration("000001_results_table.up.sql")
	if err != nil {
		log.Fatalf("failed to load migration file: %v", err)
	}

	_, err = pool.Exec(context.Background(), schemaSQL)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}
}

func TruncateResultsTable(pool *pgxpool.Pool) {
	_, err := pool.Exec(context.Background(), "TRUNCATE TABLE results")
	if err != nil {
		log.Fatalf("failed to truncate results table: %v", err)
	}
}