package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pratikdev/result-lookup/internal/models"
)

var ErrResultNotFound = errors.New("result not found")

func GetResultFromDB(pool *pgxpool.Pool, roll int, reg int, examYear int) (models.Result, error) {
	query := `SELECT roll, reg, student_name, institution_name, board_name, exam_year, gpa, is_passed
	FROM results WHERE roll = @roll AND reg = @reg AND exam_year = @examYear`
	args := pgx.NamedArgs{
		"roll": roll,
		"reg": reg,
		"examYear": examYear,
	}
	rows, err := pool.Query(context.Background(), query, args)
	if err != nil {
		rows.Close()
		return models.Result{}, err
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.Result])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrResultNotFound
		}
		return models.Result{}, err
	}

	return result, nil
}