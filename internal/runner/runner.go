package runner

import "github.com/cutekitek/rankode-runner/internal/repository/dto"

type Runner interface {
	// Syncronosly runs a test. If there are not enough resources(ram or cpu) to run a test wait for other tasks to finish
	Run(*dto.RunRequest) (*dto.RunResult, error)
}
