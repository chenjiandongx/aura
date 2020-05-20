package main

import (
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
)

var uptime = aura.NewCounter(
	"service.uptime",
	"service uptime in seconds",
	10,
	5*time.Second,
)

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(uptime)

	go func() {
		for range time.Tick(1 * time.Second) {
			uptime.Inc(1)
		}
	}()
	registry.AddReporter(reporter.DefaultStreamReporter)
	go registry.Serve("localhost:9090")
	registry.Run()
}
