//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package metrics

import (
	"sync"

	prom "github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	sandboxGuageName = "bundlelib_sandbox"
)

var (
	once      sync.Once
	collector *Collector
)

// Collector - collects bundlelib metrics
type Collector struct {
	Sandbox prom.Gauge
}

// We will never want to panic our app because of metric saving.
// Therefore, we will recover our panics here and error log them
// for later diagnosis but will never fail the app.
func recoverMetricPanic() {
	if r := recover(); r != nil {
		log.Errorf("Recovering from metric function - %v", r)
	}
}

// RegisterCollector - creates and registers bundlelib metrics collector with
// prometheus.
func RegisterCollector() {
	once.Do(func() {
		collector = &Collector{
			Sandbox: prom.NewGauge(prom.GaugeOpts{
				Name: sandboxGuageName,
				Help: "Guage of all sandbox namespaces that are active.",
			}),
		}

		err := prom.Register(collector)
		if err != nil {
			log.Errorf("unable to register collector with prometheus: %v", err)
		}
	})
}

// SandboxCreated - Counter for how many sandbox created.
func SandboxCreated() {
	defer recoverMetricPanic()
	collector.Sandbox.Inc()
}

// SandboxDeleted - Counter for how many sandbox deleted.
func SandboxDeleted() {
	defer recoverMetricPanic()
	collector.Sandbox.Dec()
}

// Describe - returns all the descriptions of the collector
func (c Collector) Describe(ch chan<- *prom.Desc) {
	c.Sandbox.Describe(ch)
}

// Collect - returns the current state of the metrics
func (c Collector) Collect(ch chan<- prom.Metric) {
	c.Sandbox.Collect(ch)
}
