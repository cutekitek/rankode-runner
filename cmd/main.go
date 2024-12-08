package main

import (
	"fmt"
	"time"

	"github.com/Qwerty10291/rankode-runner/internal/repository/dto"
	"github.com/Qwerty10291/rankode-runner/internal/runner/isolate"
)

func panicErr(err error) {
	if err != nil{
		panic(err)
	}
}

func main() {
	runner, err := isolate.NewIsolateRunner(isolate.IsolateRunnerConfig{
		MaxBoxCount:       10,
		RunnerScriptsPath: "languages",
	})
	panicErr(err)
	resp, err := runner.Run(&dto.RunRequest{
		Image:       "python3",
		Code:        `raise Exception("test")`,
		Input:       []string{
			"test",
		},
		Timeout:     time.Second,
		MemoryLimit: 100000000,
		MaxOutputSize: 1000000,
		MaxFilesSize: 100000000,
	})
	panicErr(err)
	fmt.Println(resp.Error, resp.Status, resp.Output)
}