package stats

import "runtime"

func GetCpuCoresCount() int {
	return runtime.NumCPU()
}
