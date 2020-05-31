package main

import (
	"os"
	"time"

	"github.com/chenjiandongx/aura"
)

var (
	// declare metrics
	uptime = aura.NewCounter(
		"service.uptime",
		"service uptime in seconds",
		5,
		5*time.Second,
	)
)

type MyReporter struct{}

// Custom reporter which will writes data the local file.
func (r MyReporter) Report(ch chan aura.Metric) {
	filename := "metrics.log"

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	for m := range ch {
		if _, err := f.WriteString(m.String() + "\n"); err != nil {
			panic(err)
		}
	}
}

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(uptime)

	go func() {
		for range time.Tick(1 * time.Second) {
			uptime.Inc(1)
		}
	}()

	registry.AddReporter(MyReporter{})
	registry.Run()
}
