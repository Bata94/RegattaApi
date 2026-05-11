package metrics

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/load"
)

type Metrics struct {
	CPUPercent  float64 `json:"cpu_percent"`
	RAMUsed     uint64  `json:"ram_used"`
	RAMTotal    uint64  `json:"ram_total"`
	RAMPercent  float64 `json:"ram_percent"`
	Connections int     `json:"connections"`
	Load1Min    float64 `json:"load_1min"`
	LatencyMs   float64 `json:"latency_ms"`
}

func Collect() Metrics {
	start := time.Now()

	cpuPercent, _ := cpu.Percent(0, false)

	vm, _ := mem.VirtualMemory()

	conns, _ := net.Connections("tcp")
	connections := len(conns)

	l, _ := load.Avg()

	latency := time.Since(start).Seconds() * 1000

	return Metrics{
		CPUPercent:  cpuPercent[0],
		RAMUsed:     vm.Used,
		RAMTotal:    vm.Total,
		RAMPercent:  vm.UsedPercent,
		Connections: connections,
		Load1Min:    l.Load1,
		LatencyMs:   latency,
	}
}