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
	echo = aura.NewHistogramVec(
		"http.service",
		"simple echo service",
		step,
		15*time.Second,
		[]string{"endpoint", "uri", "status"},
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
	registry.MustRegister(echo)

	go func() {
		for range time.Tick(200 * time.Millisecond) {
			echo.WithLabelValues("echo", "/api/index", "200").Observe(rand.Int63() % 600)
			echo.With(map[string]string{
				"endpoint": "echo",
				"uri":      "/api/noexists",
				"status":   "404",
			}).Observe(rand.Int63() % 600)
		}
	}()

	registry.AddReporter(reporter.DefaultStreamReporter)

	go registry.Serve("localhost:9099")
	registry.Run()
}
