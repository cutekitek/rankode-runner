package main

import (
	"fmt"

	"github.com/Qwerty10291/rankode-runner/pkg/stats"
)

func main() {
	fmt.Println(stats.GetAvailableMemory())
}
