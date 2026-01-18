package sandbox

import (
	"fmt"
	"os"

	"testing"
	"time"

	"github.com/cutekitek/rankode-runner/internal/repository/dto"
	"github.com/cutekitek/rankode-runner/internal/repository/models"
)

var sbRunner *SandboxRunner

func initSandbox() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("sandbox tests require root privileges")
	}
	cfg := SandboxRunnerConfig{
		RunnerScriptsPath:  "../../../languages",
		ContainersPoolSize: 15,
	}
	var err error
	sbRunner = NewSandboxRunner(cfg)
	if err != nil {
		return err
	}
	return sbRunner.Init()
}

func cleanupSandbox() {
	if sbRunner != nil {
		sbRunner.Close()
	}
}

func TestMain(m *testing.M) {
	if err := initSandbox(); err != nil {
		fmt.Printf("Skipping sandbox tests: %v\n", err)
		os.Exit(0)
	}
	defer cleanupSandbox()
	m.Run()
}

func TestSandboxRunner_HelloWorld(t *testing.T) {
	tests := []struct {
		language string
		code     string
		input    []string
		expected string
	}{
		{
			language: "c",
			code: `#include <stdio.h>
int main() {
    printf("Hello, World!\n");
    return 0;
}`,
			expected: "Hello, World!\n",
		},
		{
			language: "c++",
			code: `#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`,
			expected: "Hello, World!\n",
		},
		{
			language: "go",
			code: `package main
import "fmt"
func main() {
    fmt.Println("Hello, World!")
}`,
			expected: "Hello, World!\n",
		},
		{
			language: "java",
			code: `public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}`,
			expected: "Hello, World!\n",
		},
		{
			language: "js",
			code:     `console.log("Hello, World!")`,
			expected: "Hello, World!\n",
		},
		{
			language: "python3",
			code:     `print("Hello, World!")`,
			expected: "Hello, World!\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			req := &dto.RunRequest{
				Image:         tt.language,
				Code:          tt.code,
				Input:         []string{""},
				Timeout:       5000 * time.Millisecond,
				MemoryLimit:   256 * 1024 * 1024, // 256MB
				MaxFilesSize:  100 * 1024 * 1024, // 100MB
				MaxOutputSize: 1024 * 1024,       // 1MB
			}
			res, err := sbRunner.Run(req)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if res.Status != models.AttemptStatusSuccessful {
				t.Fatalf("Unexpected status: %v, error: %s", res.Status, res.Error)
			}
			if len(res.Output) != 1 {
				t.Fatalf("Expected exactly one output case, got %d", len(res.Output))
			}
			if res.Output[0].Output != tt.expected {
				t.Fatalf("Output mismatch: expected %q, got %q", tt.expected, res.Output[0].Output)
			}
		})
	}
}

func TestSandboxRunner_Verification(t *testing.T) {
	tests := []struct {
		language         string
		code             string
		verificationCode string
		expected         string
	}{
		{
			language: "python3",
			code:     "def add(a, b):\n    return a + b",
			verificationCode: `import solution
print(solution.add(2, 3))`,
			expected: "5\n",
		},
		{
			language: "js",
			code:     "function add(a, b) { return a + b; }\nmodule.exports = { add };",
			verificationCode: `const { add } = require('./main');
console.log(add(2, 3));`,
			expected: "5\n",
		},
		{
			language: "c",
			code:     "int add(int a, int b) { return a + b; }",
			verificationCode: `#include <stdio.h>
extern int add(int a, int b);
int main() { printf("%d\n", add(2, 3)); return 0; }`,
			expected: "5\n",
		},
		{
			language: "c++",
			code:     "int add(int a, int b) { return a + b; }",
			verificationCode: `#include <iostream>
extern int add(int a, int b);
int main() { std::cout << add(2, 3) << std::endl; return 0; }`,
			expected: "5\n",
		},
		{
			language: "go",
			code:     "package main\nfunc Add(a, b int) int { return a + b }",
			verificationCode: `package main
import "fmt"
func main() { fmt.Println(Add(2, 3)) }`,
			expected: "5\n",
		},
		{
			language:         "java",
			code:             "public class Main { public static int add(int a, int b) { return a + b; } }",
			verificationCode: `public class Verifier { public static void main(String[] args) { System.out.println(Main.add(2, 3)); } }`,
			expected:         "5\n",
		},
		{
			language: "sh",
			code:     "add() { echo $(($1 + $2)); }",
			verificationCode: `. ./code.sh
add 2 3`,
			expected: "5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			req := &dto.RunRequest{
				Image:            tt.language,
				Code:             tt.code,
				VerificationCode: tt.verificationCode,
				Timeout:          5000 * time.Millisecond,
				MemoryLimit:      256 * 1024 * 1024,
				MaxFilesSize:     100 * 1024 * 1024,
				MaxOutputSize:    1024 * 1024,
			}
			res, err := sbRunner.Run(req)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if res.Status != models.AttemptStatusSuccessful {
				t.Fatalf("Unexpected status: %v, error: %s", res.Status, res.Error)
			}
			if len(res.Output) != 1 {
				t.Fatalf("Expected exactly one output case, got %d", len(res.Output))
			}
			if res.Output[0].Output != tt.expected {
				t.Fatalf("Output mismatch: expected %q, got %q", tt.expected, res.Output[0].Output)
			}
		})
	}
}

func TestSandboxRunner_Input(t *testing.T) {
	tests := []struct {
		language string
		code     string
		input    []string
		expected string
	}{
		{
			language: "c",
			code: `#include <stdio.h>
int main() {
    int n;
    scanf("%d", &n);
    printf("%d\n", n * 2);
    return 0;
}`,
			input:    []string{"5"},
			expected: "10\n",
		},
		{
			language: "c++",
			code: `#include <iostream>
int main() {
    int n;
    std::cin >> n;
    std::cout << n * 2 << std::endl;
    return 0;
}`,
			input:    []string{"5"},
			expected: "10\n",
		},
		{
			language: "go",
			code: `package main
import "fmt"
func main() {
    var n int
    fmt.Scan(&n)
    fmt.Println(n * 2)
}`,
			input:    []string{"5"},
			expected: "10\n",
		},
		{
			language: "java",
			code: `import java.util.Scanner;
public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int n = sc.nextInt();
        System.out.println(n * 2);
    }
}`,
			input:    []string{"5"},
			expected: "10\n",
		},
		{
			language: "js",
			code: `const fs = require('fs');
const data = fs.readFileSync(0, 'utf8').trim();
console.log(Number(data) * 2);`,
			input:    []string{"5"},
			expected: "10\n",
		},
		{
			language: "python3",
			code: `import sys
data = sys.stdin.read().strip()
if data:
    n = int(data)
    print(n * 2)`,
			input:    []string{"5"},
			expected: "10\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			req := &dto.RunRequest{
				Image:         tt.language,
				Code:          tt.code,
				Input:         tt.input,
				Timeout:       5000 * time.Millisecond,
				MemoryLimit:   256 * 1024 * 1024,
				MaxFilesSize:  100 * 1024 * 1024,
				MaxOutputSize: 1024 * 1024,
			}
			res, err := sbRunner.Run(req)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if res.Status != models.AttemptStatusSuccessful {
				t.Fatalf("Unexpected status: %v, error: %s", res.Status, res.Error)
			}
			if len(res.Output) != 1 {
				t.Fatalf("Expected exactly one output case, got %d", len(res.Output))
			}
			if res.Output[0].Output != tt.expected {
				t.Fatalf("Output mismatch: expected %q, got %q", tt.expected, res.Output[0].Output)
			}
		})
	}
}

func TestSandboxRunner_BuildFailure(t *testing.T) {
	tests := []struct {
		language string
		code     string
	}{
		{
			language: "c",
			code:     "invalid C code",
		},
		{
			language: "c++",
			code:     "invalid C++ code",
		},
		{
			language: "go",
			code:     "invalid Go code",
		},
		{
			language: "java",
			code:     "invalid Java code",
		},
		{
			language: "js",
			code:     "invalid JS code",
		},
		{
			language: "python3",
			code:     "invalid Python code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			req := &dto.RunRequest{
				Image:         tt.language,
				Code:          tt.code,
				Input:         []string{},
				Timeout:       5000 * time.Millisecond,
				MemoryLimit:   256 * 1024 * 1024,
				MaxFilesSize:  100 * 1024 * 1024,
				MaxOutputSize: 1024 * 1024,
			}
			res, err := sbRunner.Run(req)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			hasBuildStep := tt.language == "c" || tt.language == "c++" || tt.language == "go" || tt.language == "java"
			if hasBuildStep {
				if res.Status != models.AttemptStatusBuildFailed {
					t.Fatalf("Expected build failure, got status: %v", res.Status)
				}
			}
		})
	}
}
