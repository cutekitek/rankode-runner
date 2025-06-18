package dto

import (
	"time"

	"github.com/cutekitek/rankode-runner/internal/repository/models"
)

type RunRequest struct {
	Image         string
	Code          string
	Input         []string
	Timeout       time.Duration
	MemoryLimit   int
	MaxFilesSize  int
	MaxOutputSize int
}

type RunResult struct {
	Status        models.AttemptStatus
	Error         string
	Output        []RunCaseResult
	ExecutionTime time.Duration
	MemoryUsage   int
}

type RunCaseResult struct {
	Output string
	Status models.TestCaseStatus
}
