package models

type TestCaseStatus uint8

const (
	TestCaseStatusComplete         TestCaseStatus = iota
	TestCaseStatusCompilationError TestCaseStatus = iota
	TestCaseStatusRunningError     TestCaseStatus = iota
	TestCaseStatusOutOfMemory      TestCaseStatus = iota
	TestCaseStatusTimeout          TestCaseStatus = iota
	TestCaseStatusOutputOverflow   TestCaseStatus = iota
)

type AttemptStatus uint8

const (
	AttemptStatusSuccessful    AttemptStatus = iota
	AttemptStatusBuildFailed   AttemptStatus = iota
	AttemptStatusRunFailed     AttemptStatus = iota
	AttemptStatusInternalError AttemptStatus = iota
)

type AttemptRequest struct {
	Id            int64  `json:"id"`
	Language      string `json:"language"`
	Code          string `json:"code"`
	MemoryLimit   int64  `json:"memory_limit"`
	Timeout       int64  `json:"timeout"`
	MaxOutputSize int64  `json:"max_output_size"`

	TestCases []TestCase `json:"test_cases"`
}

type TestCase struct {
	Id        int64  `db:"id" json:"id"`
	Order     int32  `db:"order" json:"order"`
	TaskId    int64  `db:"task_id" json:"task_id"`
	InputData string `db:"input" json:"input"`
}

type AttemptResponse struct {
	Id          int64         `json:"id"`
	Status      AttemptStatus `json:"status"`
	Error       string        `json:"error"`
	MemoryUsage int64         `json:"memory_usage"`
	Tests       []TestStatus  `json:"tests"`
}

type TestStatus struct {
	CaseId        int64          `json:"test_id"`
	Status        TestCaseStatus `json:"status"`
	Output        string         `json:"output"`
	ExecutionTime int64          `json:"execution_time"`
}
