package usage

import (
    "runtime"
    "fmt"
)

type TUsage struct {
    MemStats runtime.MemStats
}

func (usage *TUsage) PrintMemUsage() {
        //var m runtime.MemStats
        runtime.ReadMemStats(&usage.MemStats)
        // For info on each, see: https://golang.org/pkg/runtime/#MemStats
        fmt.Printf("Alloc = %v MiB", usage.bToMb(usage.MemStats.Alloc))
        fmt.Printf("\tTotalAlloc = %v MiB", usage.bToMb(usage.MemStats.TotalAlloc))
        fmt.Printf("\tSys = %v MiB", usage.bToMb(usage.MemStats.Sys))
        fmt.Printf("\tNumGC = %v\n", usage.MemStats.NumGC)
}

func (usage *TUsage) bToMb(b uint64) uint64 {
    return b / 1024 / 1024
}