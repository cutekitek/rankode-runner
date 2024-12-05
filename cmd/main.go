package main

import (
	"fmt"

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
		Image:       "go",
		Code:        `package main
		func main() {
		println("test")
		}`,
		Input:       []string{},
		Timeout:     100000,
		MemoryLimit: 100000,
	})
	panicErr(err)
	fmt.Println(resp.Error, resp.Status)
}