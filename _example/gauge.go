package main

import "C"
import (
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
	"github.com/shirou/gopsutil/process"
)

var (
	interval = 10 * time.Second
	cpuUsage = aura.NewGaugeVec(
		aura.BuildFQName(namespace, subsystem, "cpu.usage"),
		"CPU usage at current time",
		step,
		interval,
		[]string{"endpoint"},
	)

	memUsage = aura.NewGaugeVec(
		aura.BuildFQName(namespace, subsystem, "mem.usage"),
		"Memory usage at current time",
		step,
		interval,
		[]string{"endpoint"},
	)
)

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(cpuUsage, memUsage)
	registry.AddReporter(reporter.DefaultStreamReporter)

	go func() {
		for {
			ps, err := process.Processes()
			if err != nil {
				panic(err)
			}

			for _, p := range ps {
				pName, _ := p.Name()
				pMem, _ := p.MemoryPercent()
				pCpu, _ := p.CPUPercent()

				cpuUsage.WithLabelValues(pName).Update(pCpu)
				memUsage.WithLabelValues(pName).Update(float64(pMem))
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	//registry.Serve("127.0.0.1:9090")
	registry.Run()
}
