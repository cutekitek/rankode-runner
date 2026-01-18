package main

import (
	"fmt"
	"time"

	"github.com/cutekitek/rankode-runner/internal/repository/dto"
	"github.com/cutekitek/rankode-runner/internal/runner/sandbox"
)
const (command = "ls -lah /gocache")

func main() {
	runner := sandbox.NewSandboxRunner(sandbox.SandboxRunnerConfig{
		RunnerScriptsPath: "languages",
		ContainersPoolSize: 1,
	})
	if err := runner.Init(); err != nil {
		panic(err)
	}
	res, err := runner.Run(&dto.RunRequest{
		Image: "sh",
		Code: command,
		Input: []string{""},
		Timeout: time.Hour,
		MemoryLimit: 1024*1024*1024,
		MaxFilesSize: 1024*1024*1024,
		MaxOutputSize: 1024*1024*1024,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(command)
	fmt.Println(res.Output)
}