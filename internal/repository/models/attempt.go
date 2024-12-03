package models

type AttemptStatus int8

const (
	AttemptStatusCreated AttemptStatus = iota
	AttemptStatusTesting AttemptStatus = iota
	AttemptStatusComplete AttemptStatus = iota
	AttemptStatusWrongAnswer AttemptStatus = iota
	AttemptStatusCompilationError AttemptStatus = iota
	AttemptStatusOutOfMemory AttemptStatus = iota
	AttemptStatusTimeout AttemptStatus = iota
)

type TestStatus struct {
	TestId int64 `json:"test_id"`
	Status AttemptStatus `json:"status"`
	Output string `json:"output"`
	ExecutionTime int64 `json:"execution_time"`
}

type Attempt struct {
	Id int64
	TaskId int64
	Status AttemptStatus
	LanguageId int
	Code string
	Tests []TestStatus 
}