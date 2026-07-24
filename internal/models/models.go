package models

// all go types are string cause we receiving the value as query params
type ResultRequest struct {
	Roll int `json:"roll" db:"roll" validate:"required"`
	Reg int `json:"reg" db:"reg" validate:"required"`
	ExamYear int `json:"exam_year" db:"exam_year" validate:"required,gt=2000,lt=2200"`
}

type ResultResponse struct {
	Roll int `json:"roll" db:"roll"`
	Reg int `json:"reg" db:"reg"`
	StudentName string `json:"student_name" db:"student_name"`
	InstitutionName string `json:"institution_name" db:"institution_name"`
	BoardName string `json:"board_name" db:"board_name"`
	ExamYear int `json:"exam_year" db:"exam_year"`
	GPA float32 `json:"gpa" db:"gpa"`
	IsPassed bool `json:"is_passed" db:"is_passed"`
}

type Result struct {
	ID int `json:"-" db:"id"`
	Roll int `json:"roll" db:"roll"`
	Reg int `json:"reg" db:"reg"`
	StudentName string `json:"student_name" db:"student_name"`
	InstitutionName string `json:"institution_name" db:"institution_name"`
	BoardName string `json:"board_name" db:"board_name"`
	ExamYear int `json:"exam_year" db:"exam_year"`
	GPA float32 `json:"gpa" db:"gpa"`
	IsPassed bool `json:"is_passed" db:"is_passed"`
}