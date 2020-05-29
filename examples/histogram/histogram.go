package main

import (
	"math/rand"
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
)

const (
	step = 15
)

var (
	serviceA = aura.NewHistogram(
		"http.service.serviceA",
		"example:serviceA",
		step,
		15*time.Second,
		&aura.HistogramOpts{
			HVTypes: []aura.HistogramVType{aura.HistogramVTMin, aura.HistogramVTMax, aura.HistogramVTMean},
		},
	)

	serviceB = aura.NewHistogramVec(
		"http.service.serviceB",
		"exmaple:serviceB",
		step,
		15*time.Second,
		[]string{"endpoint", "status"},
		&aura.HistogramOpts{
			HVTypes: []aura.HistogramVType{
				aura.HistogramVTMin, aura.HistogramVTMax, aura.HistogramVTMean, aura.HistogramVTCount,
			},
			Percentiles: []float64{0.5, 0.75, 0.9, 0.99},
		},
	)
)

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(serviceA, serviceB)

	go func() {
		for range time.Tick(200 * time.Millisecond) {
			serviceA.Observe(rand.Int63() % 1000)
			serviceB.WithLabelValues("/api/index", "200").Observe(rand.Int63() % 600)
		}
	}()

	registry.AddReporter(reporter.DefaultStreamReporter)

	go registry.Serve("localhost:9099")
	registry.Run()
}
