package models

type TestCase struct {
	Id             int64  `db:"id" json:"id"`
	Order          int32  `db:"order" json:"order"`
	TaskId         int64  `db:"task_id" json:"task_id"`
	InputData      string `db:"input" json:"input"`
	ExpectedOutput string `db:"output" json:"output"`
}
