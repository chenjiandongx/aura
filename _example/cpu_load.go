package main

import (
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
	"github.com/shirou/gopsutil/load"
)

const (
	namespace = "host"
	subsystem = "cpu"
	step      = 10
)

var (
	cpuLoad1 = aura.NewDesc(
		aura.BuildFQName(namespace, subsystem, "loadavg.1"),
		"CPU load average over the last 1 minute",
		step,
		nil,
	)
	cpuLoad5 = aura.NewDesc(
		aura.BuildFQName(namespace, subsystem, "loadavg.5"),
		"load average over the last 5 minute",
		step,
		nil,
	)
	cpuLoad15 = aura.NewDesc(
		aura.BuildFQName(namespace, subsystem, "loadavg.15"),
		"load average over the last 15 minute",
		step,
		nil,
	)
)

type CPUCollector struct{}

func (c *CPUCollector) Interval() time.Duration {
	return 200 * time.Millisecond
}

func (c *CPUCollector) Describe(ch chan<- *aura.Desc) {
	ch <- cpuLoad1
	ch <- cpuLoad5
	ch <- cpuLoad15
}

func (c *CPUCollector) Collect(ch chan<- aura.Metric) {
	cpuLoad, _ := load.Avg()
	ch <- aura.MustNewConstMetric(cpuLoad1, aura.GaugeValue, cpuLoad.Load1)
	ch <- aura.MustNewConstMetric(cpuLoad5, aura.GaugeValue, cpuLoad.Load5)
	ch <- aura.MustNewConstMetric(cpuLoad15, aura.GaugeValue, cpuLoad.Load15)
}

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(&CPUCollector{})
	registry.AddReporter(reporter.DefaultStreamReporter)

	go registry.Serve("127.0.0.1:9090")
	registry.Run()
}
