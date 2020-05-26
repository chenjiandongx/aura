package main

import (
	"time"

	"github.com/chenjiandongx/aura"
	"github.com/chenjiandongx/aura/reporter"
)

var (
	uptime = aura.NewCounter(
		"service.uptime",
		"service uptime in seconds",
		5,
		5*time.Second,
	)

	reqCount = aura.NewCounterVec(
		"service.reqCount",
		"number of requests",
		5,
		5*time.Second,
		[]string{"uri", "statusCode"},
	)
)

func main() {
	registry := aura.NewRegistry(nil)
	registry.MustRegister(uptime, reqCount)

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
	registry.AddReporter(reporter.DefaultStreamReporter)

	go registry.Serve("localhost:9090")
	registry.Run()
}
