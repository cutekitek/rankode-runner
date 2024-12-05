package isolate

import "fmt"

type runFailedError struct {
	ErrorLogs  string
	StatusCode int
}

func (r *runFailedError) Error() string {
	return fmt.Sprintf("failed to run(%d)", r.StatusCode)
}
