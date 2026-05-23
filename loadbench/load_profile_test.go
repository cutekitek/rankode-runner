package loadbench

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cutekitek/rankode-runner/internal/repository/dto"
	"github.com/cutekitek/rankode-runner/internal/repository/models"
	"github.com/cutekitek/rankode-runner/internal/runner/sandbox"
)

type loadScenario struct {
	name        string
	image       string
	code        string
	inputs      []string
	pool        int
	concurrency int
	total       int
}

func TestRunnerLoadProfile(t *testing.T) {
	pythonCPU := `import sys
n = int((sys.stdin.read() or "25000").strip())
s = 0
for i in range(n):
    s += (i * i) % 97
print(s)`

	goHello := `package main

import "fmt"

func main() {
    fmt.Println("ok")
}`

	scenarios := []loadScenario{
		{name: "python_cpu_pool1_c1", image: "python3", code: pythonCPU, inputs: []string{"25000"}, pool: 1, concurrency: 1, total: 40},
		{name: "python_cpu_pool4_c4", image: "python3", code: pythonCPU, inputs: []string{"25000"}, pool: 4, concurrency: 4, total: 120},
		{name: "python_cpu_pool8_c8", image: "python3", code: pythonCPU, inputs: []string{"25000"}, pool: 8, concurrency: 8, total: 160},
		{name: "python_cpu_pool4_c16", image: "python3", code: pythonCPU, inputs: []string{"25000"}, pool: 4, concurrency: 16, total: 160},
		{name: "python_5cases_pool8_c8", image: "python3", code: pythonCPU, inputs: []string{"10000", "15000", "20000", "25000", "30000"}, pool: 8, concurrency: 8, total: 80},
		{name: "go_build_pool1_c1", image: "go", code: goHello, inputs: []string{""}, pool: 1, concurrency: 1, total: 20},
		{name: "go_build_pool4_c4", image: "go", code: goHello, inputs: []string{""}, pool: 4, concurrency: 4, total: 40},
		{name: "go_build_pool8_c8", image: "go", code: goHello, inputs: []string{""}, pool: 8, concurrency: 8, total: 40},
		{name: "go_build_pool4_c16", image: "go", code: goHello, inputs: []string{""}, pool: 4, concurrency: 16, total: 80},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			runLoadScenario(t, scenario)
		})
	}
}

func runLoadScenario(t *testing.T, scenario loadScenario) {
	runner := sandbox.NewSandboxRunner(sandbox.SandboxRunnerConfig{
		RunnerScriptsPath:  "../languages",
		ContainersPoolSize: scenario.pool,
	})
	if err := runner.Init(); err != nil {
		t.Fatalf("failed to initialize sandbox runner: %v", err)
	}
	defer runner.Close()

	req := &dto.RunRequest{
		Image:         scenario.image,
		Code:          scenario.code,
		Input:         scenario.inputs,
		Timeout:       5 * time.Second,
		MemoryLimit:   256 * 1024 * 1024,
		MaxFilesSize:  100 * 1024 * 1024,
		MaxOutputSize: 1024 * 1024,
	}

	if res, err := runner.Run(req); err != nil || res.Status != models.AttemptStatusSuccessful {
		if err != nil {
			t.Fatalf("warmup failed: %v", err)
		}
		t.Fatalf("warmup returned status %v: %s", res.Status, res.Error)
	}

	jobs := make(chan int)
	durations := make(chan time.Duration, scenario.total)
	var failures atomic.Int64
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < scenario.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				requestStart := time.Now()
				res, err := runner.Run(req)
				durations <- time.Since(requestStart)
				if err != nil || res.Status != models.AttemptStatusSuccessful {
					failures.Add(1)
				}
			}
		}()
	}

	for i := 0; i < scenario.total; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(durations)
	wall := time.Since(start)

	values := make([]time.Duration, 0, scenario.total)
	for duration := range durations {
		values = append(values, duration)
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })

	throughput := float64(scenario.total) / wall.Seconds()
	avg := average(values)
	fmt.Printf("LOAD_RESULT scenario=%s pool=%d concurrency=%d total=%d ok=%d errors=%d wall_ms=%d throughput_rps=%.2f avg_ms=%.2f min_ms=%.2f p50_ms=%.2f p95_ms=%.2f p99_ms=%.2f max_ms=%.2f\n",
		scenario.name,
		scenario.pool,
		scenario.concurrency,
		scenario.total,
		scenario.total-int(failures.Load()),
		failures.Load(),
		wall.Milliseconds(),
		throughput,
		ms(avg),
		ms(values[0]),
		ms(percentile(values, 0.50)),
		ms(percentile(values, 0.95)),
		ms(percentile(values, 0.99)),
		ms(values[len(values)-1]),
	)
}

func average(values []time.Duration) time.Duration {
	var total time.Duration
	for _, value := range values {
		total += value
	}
	return total / time.Duration(len(values))
}

func percentile(values []time.Duration, p float64) time.Duration {
	idx := int(math.Ceil(float64(len(values))*p)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(values) {
		idx = len(values) - 1
	}
	return values[idx]
}

func ms(duration time.Duration) float64 {
	return float64(duration.Microseconds()) / 1000
}
