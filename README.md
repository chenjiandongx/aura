# Aura

> 🔔 Aura is a SDK for the monitoring system written in Go with love.

[![GoDoc](https://godoc.org/github.com/chenjiandongx/aura?status.svg)](https://godoc.org/github.com/chenjiandongx/aura)
[![Go Report Card](https://goreportcard.com/badge/github.com/chenjiandongx/aura)](https://goreportcard.com/report/github.com/chenjiandongx/aura)
[![License](https://img.shields.io/badge/License-apache-brightgreen.svg)](https://www.apache.org/)

## 🎬 Overview

☁️ 在云原生时代，以 [Prometheus](https://prometheus.io) 为中心的监控生态已经逐渐完善，社区也出现了大量的中间件，数据库以及各种基础组件的 exporter，Prometheus 官方也给出了维护了一份 exporter 列表 [instrumenting/exporters](https://prometheus.io/docs/instrumenting/exporters)。

但是 Prometheus 的缺点和它的优点一样明显，缺少高可用的集群方案。想了解 Prometheus 和监控系统的同学可阅读 [Prometheus 折腾笔记](https://github.com/chenjiandongx/prometheus101) 系列文章。目前开源的高可用企业级的监控方案有两个，小米的 [falcon-plus](https://github.com/open-falcon/falcon-plus) 和滴滴出行的 [nightingale](https://github.com/didi/nightingale)，后者是前者的优化增强版。国内的不少公司（比如我司 🐶）的监控方案都或多或少参考了 falcon 的设计架构，falcon 的官网也维护了一份 [企业用户列表](http://book.open-falcon.org/zh_0_2/contributing.html)。

falcon 的设计架构决定了它强悍的性能及良好的可扩展性，具体关于其相关信息可参考 [官网介绍](http://book.open-falcon.org/zh_0_2/intro/)。falcon 的 slogan：
> *open-falcon 的目标是做最开放、最好用的互联网企业级监控产品。*

目前 falcon 已经不再维护 😔，可能它已经完成了它的历史使命吧，提供一套完整监控系统的构建方案；不过 nightingale 接住了 falcon 手中的接力棒，为开源社区的监控领域又注入了新的活力 😌。虽然如此，但是 falcon/nightingale 所构建的生态也仍旧不完善，缺少像 Prometheus 生态的各类数据采集器（exporter）。以 Prometheus 为中心的采集器都是通过 **暴露 HTTP 端口** 来让服务端采集，是一种 **Pull** 模式，而 falcon 体系是采用 **Push** 模式，**客户端主动上报**。数据采集形态的不同应该是 Prometheus 和 falcon 的最大差异点。

## 💡 Idea

🤔 如果有一种方案，能够以比较低的开发成本，将 Prometheus 的 exporter 转为换 falcon 的 collector，那样的话 falcon 的生态就会变得丰富多彩。

* Metric 是监控体系中的重要概念，一个 metric 代表着一个监控项。Java 有一个优秀的 metric 相关的开源库 [dropwizard/metrics](https://github.com/dropwizard/metrics)，同时也有开发者基于该库开发了一个 Golang 版本 [rcrowley/go-metrics](https://github.com/rcrowley/go-metrics)，关于这两个库的更多信息，可移步至项目其地址。

* Prometheus 本身在提供服务端的同时，也开发不同语言的 SDK 客户端，如 Golang 版本 [prometheus/client_golang](https://github.com/proemtheus/client_golang)。

当 rcrowley/go-metrics 遇上 prometheus/client_golang，[Aura](https://github.com/chenjiandongx/aura) 就出现啦 🥺。如果你使用过 Prometheus 的 SDK，那你将会对 Aura 提供的接口非常熟悉。Aura 的目标是成为 falcon 体系的客户端 SDK。

## 🔰 Installation

```shell
$ go get -u github.com/chenjiandongx/aura/...
```

## 🔖 Metric

Aura 标准 Metric 结构，沿用了 falcon 的设计。
```golang
type Metric struct {
	Endpoint  string
	Metric    string
	Step      uint32
	Value     interface{}
	Type      ValueType
	Labels    map[string]string
	Timestamp int64
}
```

### * Counter

Counter 单调递增，违反单调性时重置为 0。可以用于统计某些事件出现的次数，或者服务的 uptime。
```golang
type Counter interface {
	Collector

	Clear()
	Count() int64
	Dec(int64)
	Inc(int64)
}
```

### * Gauge

Gauge 记录瞬时值，可以用于记录系统当下时刻的状态，比如 CPU 使用率，使用内存大小，网络 IO 情况。
```golang
type Gauge interface {
	Collector

	Update(float64)
	Value() float64
}
```

### * Histogram

Histogram 主要用于表示一段时间范围内对数据进行采样，并能够对其指定区间以及总数进行统计，通常它采集的数据展示为直方图。
```golang
type Histogram interface {
	Collector

	Observe(int64)
}
```

### * Timer

Timer 主要用于统计一段代码逻辑或一次事件的耗时分布。
```golang
type Timer interface {
	Collector

	Time(func())
	Update(time.Duration)
}
```

## 📝 Usage

### Registry

Registry 负责注册和管理 Collectors 的生命周期。

```golang
// RegistryOpts 用于指定 Metrics 和 Desc channel 的缓存大小。
// 一般情况下不需要调整，如果采集指标量比较大的话，可以将 CapMetricChan 值设置大一点。
type RegistryOpts struct {
	CapMetricChan int // default 2500
	CapDescChan   int // default 20
}

func NewRegistry(opts *RegistryOpts) *Registry
```

### Collector 基本用法

```golang
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
	// 使用 aura.NewDesc 声明采集的指标
	// NewDesc(fqName, help string, step uint32, labelKeys []string) *Desc 
	// * fqName: 指标名称
 	// * help: 指标描述或者介绍（可为空）
	// * step: 指标步长
	// * labelkeys: 指标 label keys。
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

// Interval 实现了 aura.Collector 接口。声明采集时间。
func (c *CPUCollector) Interval() time.Duration {
	return 2 * time.Second
}

// Describe 实现了 aura.Collector 接口。注册指标。
func (c *CPUCollector) Describe(ch chan<- *aura.Desc) {
	ch <- cpuLoad1
	ch <- cpuLoad5
	ch <- cpuLoad15
}

// Describe 实现了 aura.Collector 接口。指标具体采集逻辑。
func (c *CPUCollector) Collect(ch chan<- aura.Metric) {
	cpuLoad, _ := load.Avg()
	ch <- aura.MustNewConstMetric(cpuLoad1, aura.GaugeValue, cpuLoad.Load1)
	ch <- aura.MustNewConstMetric(cpuLoad5, aura.GaugeValue, cpuLoad.Load5)
	ch <- aura.MustNewConstMetric(cpuLoad15, aura.GaugeValue, cpuLoad.Load15)
}

func main() {
	// (1) 创建一个 Rigistry 对象
	registry := aura.NewRegistry(nil)
	// (2) 注册 Collector
	registry.MustRegister(&CPUCollector{})
	// (3) 注册 Reporter
	// reporter 负责将 metrics 输送到任意后端，开发者可自行为 registry 提供定制化后端
	// reporter.DefaultStreamReporter 会将采集的指标输出到 stdout
	registry.AddReporter(reporter.DefaultStreamReporter)

	// 可选项：Serve 将会启动一个 HTTP 服务用于提供 collector 本身运行的信息。
	go registry.Serve("127.0.0.1:9099")
	// (4) 开始采集指标
	registry.Run()
}
```

**运行结果**

```shell
~/project/golang/src/github.com/chenjiandongx/aura/examples/desc 🤔 go run .
{Endpoint: Metric:host.cpu.loadavg.15 Step:10 Value:2.01318359375 Type:Gauge Labels:map[] Timestamp:1590776801}
{Endpoint: Metric:host.cpu.loadavg.15 Step:10 Value:2.01318359375 Type:Gauge Labels:map[] Timestamp:1590776803}
{Endpoint: Metric:host.cpu.loadavg.15 Step:10 Value:2.01318359375 Type:Gauge Labels:map[] Timestamp:1590776805}
{Endpoint: Metric:host.cpu.loadavg.1 Step:10 Value:1.60791015625 Type:Gauge Labels:map[] Timestamp:1590776807}
{Endpoint: Metric:host.cpu.loadavg.5 Step:10 Value:2.02587890625 Type:Gauge Labels:map[] Timestamp:1590776801}
{Endpoint: Metric:host.cpu.loadavg.5 Step:10 Value:2.02587890625 Type:Gauge Labels:map[] Timestamp:1590776803}
{Endpoint: Metric:host.cpu.loadavg.1 Step:10 Value:1.748046875 Type:Gauge Labels:map[] Timestamp:1590776805}
{Endpoint: Metric:host.cpu.loadavg.15 Step:10 Value:2.0009765625 Type:Gauge Labels:map[] Timestamp:1590776807}
...
```

**Collector 指标及运行状态**

```shell
~/project/golang/src/github.com/chenjiandongx/aura 🤔 curl -s http://localhost:9099/-/metadata | jq
[
  {
    "metric": "host.cpu.loadavg.1",
    "help": "CPU load average over the last 1 minute",
    "step": 10
  },
  {
    "metric": "host.cpu.loadavg.5",
    "help": "load average over the last 5 minute",
    "step": 10
  },
  {
    "metric": "host.cpu.loadavg.15",
    "help": "load average over the last 15 minute",
    "step": 10
  }
]
~/project/golang/src/github.com/chenjiandongx/aura 🤔 curl -s http://localhost:9099/-/stats | jq
{
  "metricsChanCap": 2500,
  "metricsChanLen": 0
}
```


### 客户端埋点形式

```golang
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
	// 声明采集指标
	echo = aura.NewHistogramVec(
		"http.service",
		"simple echo service",
		step,
		15*time.Second,
		[]string{"endpoint", "uri", "status"},
		// 直方图上报数据如果指定了 HVTypes/Percentiles 那上报就是计算后的指标
		// 计算后的指标形式
		// http.service.min
		// http.service.max
		// http.service.mean
		// http.service.count
		// http.service.0.50
		// http.service.0.75
		// http.service.0.90
		// http.service.0.99
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
```

**运行结果**

```shell
~/project/golang/src/github.com/chenjiandongx/aura/examples/histogram 🤔 go run .
{Endpoint:echo Metric:http.service.max Step:15 Value:590 Type:Gauge Labels:map[endpoint:echo status:200 uri:/api/index] Timestamp:1590778743}
{Endpoint:echo Metric:http.service.0.75 Step:15 Value:460.5 Type:Gauge Labels:map[endpoint:echo status:200 uri:/api/index] Timestamp:1590778743}
{Endpoint:echo Metric:http.service.0.99 Step:15 Value:590 Type:Gauge Labels:map[endpoint:echo status:200 uri:/api/index] Timestamp:1590778743}
{Endpoint:echo Metric:http.service.count Step:15 Value:20 Type:Gauge Labels:map[endpoint:echo status:404 uri:/api/noexists] Timestamp:1590778743}
{Endpoint:echo Metric:http.service.0.50 Step:15 Value:325.5 Type:Gauge Labels:map[endpoint:echo status:404 uri:/api/noexists] Timestamp:1590778743}
{Endpoint:echo Metric:http.service.0.75 Step:15 Value:460.5 Type:Gauge Labels:map[endpoint:echo status:404 uri:/api/noexists] Timestamp:1590778743}
{Endpoint:echo Metric:http.service.0.90 Step:15 Value:572.4000000000001 Type:Gauge Labels:map[endpoint:echo status:404 uri:/api/noexists] Timestamp:1590778743}
{Endpoint:echo Metric:http.service.0.99 Step:15 Value:590 Type:Gauge Labels:map[endpoint:echo status:404 uri:/api/noexists] Timestamp:1590778743}
...
```

### 自定义 Reporter

```golang
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
```

**运行结果**

```shell
~/project/golang/src/github.com/chenjiandongx/aura/examples/reporter 🤔 tail -f metrics.log
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945775, Step:5, Value:1, Tags:map[]>
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945778, Step:5, Value:5, Tags:map[]>
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945783, Step:5, Value:10, Tags:map[]>
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945788, Step:5, Value:15, Tags:map[]>
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945793, Step:5, Value:19, Tags:map[]>
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945798, Step:5, Value:25, Tags:map[]>
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945803, Step:5, Value:29, Tags:map[]>
<Metadata Endpoint:, Metric:service.uptime, Type:Counter Timestamp:1590945808, Step:5, Value:34, Tags:map[]>
...
```

Aura 提供了一些示例位于 [examples](https://github.com/chenjiandongx/aura/tree/master/examples) 文件夹。同时也基于 [prometheus/memcached_exporter](https://github.com/prometheus/memcached_exporter) 开发了 [memcached-collector](https://github.com/chenjiandongx/memcached-collector)，作为一个标准 collector 写法供使用的同学参考。

### 📃 License

[Apache License v2](https://github.com/chenjiandongx/aura/LICENSE)
