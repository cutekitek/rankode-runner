package isolate

import (
	"errors"
	"fmt"
)

type runFailedError struct {
	ErrorLogs  string
	StatusCode int
}

func (r *runFailedError) Error() string {
	return fmt.Sprintf("failed to run(%d)", r.StatusCode)
}

var OutputOverflowErr = errors.New("output is too big")
