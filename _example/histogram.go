package main

import (
	"math/rand"
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
)

var service = aura.NewHistogram(
	"service.stats",
	"stats of service",
	30,
	10*time.Second,
	&aura.HistogramOpts{
		HVTypes:     []aura.HistogramVType{aura.HVTMin, aura.HVTMax, aura.HVTMean},
		Percentiles: []float64{0.5, 0.75, 0.9, 0.99},
	},
)

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(service)

	go func() {
		for range time.Tick(300 * time.Millisecond) {
			service.Observe(rand.Int63() % 1000)
		}
	}()

	registry.AddReporter(reporter.DefaultStreamReporter)
	go registry.Serve("localhost:9090")
	registry.Run()
}
