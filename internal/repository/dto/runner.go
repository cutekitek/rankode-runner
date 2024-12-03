package dto

import (
	"time"
)

type RunRequest struct {
	Image     string
	Code      string
	Input     []string
	Timeout   time.Duration
	// В байтах
	MemoryLimit int
}

type RunResult struct {
	Output        []string
	ExecutionTime int64
	MemoryUsage   int
}
