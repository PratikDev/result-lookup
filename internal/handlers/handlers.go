package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pratikdev/result-lookup/internal/database"
	"github.com/pratikdev/result-lookup/internal/models"
	redisClient "github.com/pratikdev/result-lookup/internal/redis"
	"github.com/pratikdev/result-lookup/internal/utils"
	"github.com/redis/go-redis/v9"
)

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}

func GetResult(w http.ResponseWriter, r *http.Request, fbDbPool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger, validate *validator.Validate) {
	queryParams := r.URL.Query()

	roll, err := strconv.Atoi(queryParams.Get("roll"))
	if err != nil {
		http.Error(w, "invalid roll number", http.StatusBadRequest)
    return
	}

	reg, err := strconv.Atoi(queryParams.Get("reg"))
	if err != nil {
		http.Error(w, "invalid reg number", http.StatusBadRequest)
    return
	}

	examYear, err := strconv.Atoi(queryParams.Get("exam_year"))
	if err != nil {
		http.Error(w, "invalid exam year", http.StatusBadRequest)
    return
	}

	requestBody := models.ResultRequest{
		Roll: roll,
		Reg: reg,
		ExamYear: examYear,
	}

	// validate
	err = validate.Struct(requestBody)
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

	resultKey := fmt.Sprintf("result:%d:%d:%d", requestBody.Roll, requestBody.Reg, requestBody.ExamYear)
	result, err := redisClient.Get(rdb, logger, resultKey)
	if err != nil {
		code := http.StatusInternalServerError
		errorMsg := err.Error()

		if errors.Is(err, redis.Nil) {
			code = http.StatusNotFound
			errorMsg = "didn't find result"
		} else {
			/* postgres fallback */


			// we shouldn't kill the process
			// if the fallback db connection failed
			if fbDbPool == nil {
				http.Error(w, errorMsg, code)
				return
			}

			resultModel, err := database.GetResultFromDB(fbDbPool, requestBody.Roll, requestBody.Reg, requestBody.ExamYear)
			if err != nil {
				errorMsg = err.Error()
				if errors.Is(err, database.ErrResultNotFound) {
					code = http.StatusNotFound
				}
				http.Error(w, errorMsg, code)
				return
			}

			result, err = utils.ToJSON(resultModel)
			if err != nil {
				logger.Error("result model to json process failed","error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// fallback succeeded
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(result))
			return
		}

		http.Error(w, errorMsg, code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}