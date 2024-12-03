package runner

import "github.com/Qwerty10291/rankode-runner/internal/repository/dto"

type Runner interface {
	Run(*dto.RunRequest) (*dto.RunResult, error)
}
