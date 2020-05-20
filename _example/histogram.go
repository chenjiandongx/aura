package main

import (
	"math/rand"
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
)

var (
	serviceEcho = aura.NewHistogram(
		"http.service.echo",
		"http service stats",
		15,
		15*time.Second,
		&aura.HistogramOpts{
			HVTypes: []aura.HistogramVType{aura.HVTMin, aura.HVTMax, aura.HVTMean},
		},
	)

	serviceHome = aura.NewHistogramVec(
		"http.service.home",
		"http service stats",
		15,
		15*time.Second,
		[]string{"endpoint"},
		&aura.HistogramOpts{
			HVTypes:     []aura.HistogramVType{aura.HVTMin, aura.HVTMax, aura.HVTMean, aura.HVTCount},
			Percentiles: []float64{0.5, 0.75, 0.9, 0.99},
		},
	)
)

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(serviceEcho, serviceHome)

	go func() {
		for range time.Tick(200 * time.Millisecond) {
			serviceEcho.Observe(rand.Int63() % 1000)
			serviceHome.WithLabelValues("/api/index").Observe(rand.Int63() % 600)
		}
	}()

	registry.AddReporter(reporter.DefaultStreamReporter)
	
	go registry.Serve("localhost:9090")
	registry.Run()
}
