package main

import (
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
)

var (
	// declare metrics
	uptime = aura.NewCounter(
		"service.uptime",
		"service uptime in seconds",
		5,
		5*time.Second,
	)

	reqCount = aura.NewCounterVec(
		"service.reqCount",
		"count of requests",
		5,
		5*time.Second,
		[]string{"uri", "statusCode"},
	)
)

func main() {
	// (1) create a new registry
	registry := aura.NewRegistry(nil)
	// (2) register collectors.
	// *Counter* is a Collector which has implemented the aura.Collector.
	registry.MustRegister(uptime, reqCount)

	// (3) do your collecting stuffs
	go func() {
		for range time.Tick(1 * time.Second) {
			uptime.Inc(1)
		}
	}()

	go func() {
		for range time.Tick(200 * time.Millisecond) {
			reqCount.WithLabelValues("/api", "200").Inc(1)
			reqCount.With(map[string]string{"uri": "/index", "statusCode": "400"}).Inc(1)
		}
	}()

	// (4) add reporter for reporting to any backend you want.
	// reporter.DefaultStreamReporter here will prints metrics collected to the stdout.
	registry.AddReporter(reporter.DefaultStreamReporter)

	// Optional: Serve will run a HTTP server which exports more information about collector itself.
	go registry.Serve("localhost:9099")

	// (5) run forever to collect metrics you declared.
	registry.Run()
}
