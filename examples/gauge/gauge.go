package gauge

import (
	"github.com/chenjiandongx/aura/examples/desc"
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
	"github.com/shirou/gopsutil/process"
)

var (
	interval = 10 * time.Second
	cpuUsage = aura.NewGauge(
		aura.BuildFQName(desc.namespace, desc.subsystem, "cpu.usage"),
		"CPU usage at current time",
		desc.step,
		interval,
	)

	memUsage = aura.NewGaugeVec(
		aura.BuildFQName(desc.namespace, desc.subsystem, "mem.usage"),
		"Memory usage at current time",
		desc.step,
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

				cpuUsage.Update(pCpu)
				memUsage.WithLabelValues(pName).Update(float64(pMem))
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	go registry.Serve("127.0.0.1:9090")
	registry.Run()
}
