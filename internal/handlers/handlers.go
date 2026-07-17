package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/pratikdev/result-lookup/internal/models"
	redisClient "github.com/pratikdev/result-lookup/internal/redis"
	"github.com/redis/go-redis/v9"
)

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}

func GetResult(w http.ResponseWriter, r *http.Request, rdb *redis.Client, logger *slog.Logger, validate *validator.Validate) {
	queryParams := r.URL.Query()

	requestBody := models.ResultRequest{
		Roll: queryParams.Get("roll"),
		Reg: queryParams.Get("reg"),
		ExamYear: queryParams.Get("exam_year"),
	}

	// validate
	err := validate.Struct(requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check publish gate
	isPublished, err := redisClient.Get(rdb, logger, "result:published")
	if err != nil {
		code := http.StatusInternalServerError
		errorMsg := err.Error()
		
		if errors.Is(err, redis.Nil) {
			code = http.StatusServiceUnavailable
			errorMsg = "result not published yet"
		}

		http.Error(w, errorMsg, code)
		return
	}

	if isPublished == "0" {
		http.Error(w, "result not published yet", http.StatusServiceUnavailable)
		return
	}

	resultKey := fmt.Sprintf("result:%s:%s:%s", requestBody.Roll, requestBody.Reg, requestBody.ExamYear)
	result, err := redisClient.Get(rdb, logger, resultKey)
	if err != nil {
		code := http.StatusInternalServerError
		errorMsg := err.Error()

		if errors.Is(err, redis.Nil) {
			code = http.StatusNotFound
			errorMsg = "didn't find result"
		}

		http.Error(w, errorMsg, code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}