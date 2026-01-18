package benchmarks

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/cutekitek/rankode-runner/internal/repository/dto"
	"github.com/cutekitek/rankode-runner/internal/repository/models"
	"github.com/cutekitek/rankode-runner/internal/runner/sandbox"
)

var (
	runner *sandbox.SandboxRunner
)

func initSandbox() error {
	cfg := sandbox.SandboxRunnerConfig{
		RunnerScriptsPath:  "../languages",
		ContainersPoolSize: 1,
	}
	var err error
	runner = sandbox.NewSandboxRunner(cfg)
	if err != nil {
		return err
	}
	return runner.Init()
}

func cleanupSandbox() {
	if runner != nil {
		runner.Close()
	}
}

func TestMain(m *testing.M) {
	if err := initSandbox(); err != nil {
		log.Fatalf("Failed to init sandbox runner: %v", err)
	}
	defer cleanupSandbox()
	m.Run()
}

func BenchmarkGoHelloWorld(b *testing.B) {
	req := &dto.RunRequest{
		Image: "go",
		Code: `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`,
		Input:         []string{
			"",
		},
		Timeout:       5000 * time.Millisecond,
		MemoryLimit:   256 * 1024 * 1024, // 256MB
		MaxFilesSize:  100 * 1024 * 1024, // 100MB
		MaxOutputSize: 1024 * 1024,       // 1MB
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := runner.Run(req)
		fmt.Println(res.Output)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
		if res.Status != models.AttemptStatusSuccessful {
			b.Fatalf("Unexpected status: %v %s", res.Status, res.Error)
		}
	}
}

func BenchmarkGoFibonacci(b *testing.B) {
	code := `package main

import "fmt"

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	fmt.Println(fib(20))
}`
	req := &dto.RunRequest{
		Image:         "go",
		Code:          code,
		Input:         []string{},
		Timeout:       5000 * time.Millisecond,
		MemoryLimit:   256 * 1024 * 1024,
		MaxFilesSize:  100 * 1024 * 1024,
		MaxOutputSize: 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := runner.Run(req)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
		if res.Status != models.AttemptStatusSuccessful {
			b.Fatalf("Unexpected status: %v", res.Status)
		}
	}
}

func BenchmarkGoMultipleInputs(b *testing.B) {
	code := `package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	data, _ := os.ReadFile("/dev/stdin")
	n, _ := strconv.Atoi(string(data))
	fmt.Println(n * 2)
}`
	inputs := []string{"5", "10", "15", "20", "25"}
	req := &dto.RunRequest{
		Image:         "go",
		Code:          code,
		Input:         inputs,
		Timeout:       5000 * time.Millisecond,
		MemoryLimit:   256 * 1024 * 1024,
		MaxFilesSize:  100 * 1024 * 1024,
		MaxOutputSize: 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := runner.Run(req)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
		if res.Status != models.AttemptStatusSuccessful {
			b.Fatalf("Unexpected status: %v", res.Status)
		}
	}
}

func BenchmarkPythonHelloWorld(b *testing.B) {
	req := &dto.RunRequest{
		Image:         "python3",
		Code:          `print("Hello, World!")`,
		Input:         []string{},
		Timeout:       5000 * time.Millisecond,
		MemoryLimit:   256 * 1024 * 1024,
		MaxFilesSize:  100 * 1024 * 1024,
		MaxOutputSize: 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := runner.Run(req)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
		if res.Status != models.AttemptStatusSuccessful {
			b.Fatalf("Unexpected status: %v", res.Status)
		}
	}
}

func BenchmarkPythonFibonacci(b *testing.B) {
	code := `def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)

print(fib(20))`
	req := &dto.RunRequest{
		Image:         "python3",
		Code:          code,
		Input:         []string{},
		Timeout:       5000 * time.Millisecond,
		MemoryLimit:   256 * 1024 * 1024,
		MaxFilesSize:  100 * 1024 * 1024,
		MaxOutputSize: 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := runner.Run(req)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
		if res.Status != models.AttemptStatusSuccessful {
			b.Fatalf("Unexpected status: %v", res.Status)
		}
	}
}

func BenchmarkPythonInputProcessing(b *testing.B) {
	code := `import sys
data = sys.stdin.read().strip()
if data:
    n = int(data)
    print(n * 2)
else:
    print("no input")`
	inputs := []string{"5", "10", "15", "20", "25"}
	req := &dto.RunRequest{
		Image:         "python3",
		Code:          code,
		Input:         inputs,
		Timeout:       5000 * time.Millisecond,
		MemoryLimit:   256 * 1024 * 1024,
		MaxFilesSize:  100 * 1024 * 1024,
		MaxOutputSize: 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := runner.Run(req)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
		if res.Status != models.AttemptStatusSuccessful {
			b.Fatalf("Unexpected status: %v", res.Status)
		}
	}
}
