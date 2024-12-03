package docker

import (
	"sync"

	"github.com/Qwerty10291/rankode-runner/internal/repository/dto"
	"github.com/Qwerty10291/rankode-runner/internal/runner"
	"github.com/docker/docker/client"
)

type DockerRunner struct {
	cli *client.Client
	availableCPUs chan int
	mu *sync.Mutex
}

type DockerRunnerConfig struct {
	// How many cpu cores will be used
	CpuCores int
	// how many tasks can be run on a single core
	TasksPerCpu int
}

func (d *DockerRunner) Run(*dto.RunRequest) (*dto.RunResult, error) {
	// wait for cpu core
	core := <- d.availableCPUs
	defer func() {
		d.availableCPUs <- core
	}()
	
	return nil, nil
}


func NewDockerRunner(cfg DockerRunnerConfig) (runner.Runner, error) {
	cli, err := client.NewClientWithOpts()
	cpusQueue := make(chan int, cfg.CpuCores * cfg.TasksPerCpu)
	for i := 0; i < cfg.CpuCores; i++ {
		for j := 0; j < cfg.TasksPerCpu; j++ {
			cpusQueue <- i;
		}
	}

	return &DockerRunner{
		cli: cli,
		mu: new(sync.Mutex),
		availableCPUs: cpusQueue,
	}, err
}


