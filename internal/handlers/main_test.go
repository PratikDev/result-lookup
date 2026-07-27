package handlers

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pratikdev/result-lookup/internal/testutils"
	"github.com/redis/go-redis/v9"
)

var testValidator *validator.Validate
var testFbDbPool *pgxpool.Pool
var testMR *miniredis.Miniredis
var testRDB *redis.Client
var	testLogger *slog.Logger

func TestMain(m *testing.M) {
	// Global Setup
	fmt.Println("Setting up global resources...")

	// load test env vars
	err := godotenv.Load("../../.env.test")
  if err != nil {
    log.Fatal("Error loading .env.test file")
		return
  }

	// validator setup
	testValidator = validator.New()

	// fallback db pool setup
	testFbDbPool = testutils.InitTestDB()
	testutils.SchemaSetup(testFbDbPool)

	// miniredis setup
	testMR, err = miniredis.Run()
	if err != nil {
		log.Fatalf("failed to started miniredis: %v", err)
		return
	}
	testRDB = redis.NewClient(&redis.Options{
		Addr: testMR.Addr(),
	})

	// logger setup
	handler := slog.NewJSONHandler(os.Stderr, nil)
	testLogger = slog.New(handler)

	// Run all unit tests in this package
	exitCode := m.Run()

	// Global Teardown
	fmt.Println("Cleaning up global resources...")
	testMR.Close()

	// Exit with the correct status code
	os.Exit(exitCode)
}