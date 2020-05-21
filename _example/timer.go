package main

import (
	"math/rand"
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
)

var (
	timerA = aura.NewTimer(
		"host.timerA",
		"it's the timerA to record something",
		15,
		15*time.Second,
		&aura.TimerOpts{
			HVTypes: []aura.TimerVType{aura.TimerVTMin, aura.TimerVTMax, aura.TimerVTStdDev},
		},
	)

	timerB = aura.NewTimerVec(
		"host.timerB",
		"it's the timerB to record something",
		15,
		15*time.Second,
		[]string{"endpoint"},
		&aura.TimerOpts{
			HVTypes:     []aura.TimerVType{aura.TimerVTMin, aura.TimerVTStdDev},
			Percentiles: []float64{0.5, 0.75, 0.9, 0.99},
		},
	)
)

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(timerA, timerB)

	go func() {
		for range time.Tick(200 * time.Millisecond) {
			timerA.Update(time.Duration(rand.Int63()%1000) * time.Millisecond)
			timerB.WithLabelValues("/api/index").Update(time.Duration(rand.Int63()%600) * time.Millisecond)
		}
	}()

	registry.AddReporter(reporter.DefaultStreamReporter)

	go registry.Serve("localhost:9090")
	registry.Run()
}
