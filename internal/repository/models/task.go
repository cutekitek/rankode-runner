package models

type Task struct {
	Id int64 `db:"id" json:"id"`
	TestCases []TestCase `json:"test_cases"`
}