//
// Copyright (c) 2017 Red Hat, Inc.
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
	logging "github.com/op/go-logging"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subsystem = "asb"
)

var (
	sandbox = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "sandbox",
			Help:      "Gauge of all sandbox namespaces that are active.",
		})

	specsLoaded = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "specs_loaded",
			Help:      "Specs loaded from registries, partitioned by registry name.",
		}, []string{"registry_name"})

	specsReset = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "specs_reset",
			Help:      "Counter of how many times the specs have been reset.",
		})

	provisionJob = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "provision_jobs",
			Help:      "How many provision jobs are actively in the buffer.",
		})

	deprovisionJob = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "deprovision_jobs",
			Help:      "How many deprovision jobs are actively in the buffer.",
		})

	updateJob = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "update_jobs",
			Help:      "How many update jobs are actively in the buffer.",
		})

	bindingJob = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "binding_jobs",
			Help:      "How many binding jobs are actively in the buffer.",
		})

	requests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "actions_requested",
			Help:      "How many actions have been made.",
		}, []string{"action"})

	log = logging.MustGetLogger("metrics")
)

func init() {
	prometheus.MustRegister(sandbox)
	prometheus.MustRegister(specsLoaded)
	prometheus.MustRegister(specsReset)
	prometheus.MustRegister(provisionJob)
	prometheus.MustRegister(deprovisionJob)
	prometheus.MustRegister(updateJob)
	prometheus.MustRegister(requests)
}

// We will never want to panic our app because of metric saving.
// Therefore, we will recover our panics here and error log them
// for later diagnosis but will never fail the app.
func recoverMetricPanic() {
	if r := recover(); r != nil {
		log.Errorf("Recovering from metric function - %v", r)
	}
}

// SandboxCreated - Counter for how many sandbox created.
func SandboxCreated() {
	defer recoverMetricPanic()
	sandbox.Inc()
}

// SandboxDeleted - Counter for how many sandbox deleted.
func SandboxDeleted() {
	defer recoverMetricPanic()
	sandbox.Dec()
}

// SpecsLoaded - Will add the count of specs. (The value can be negative,
// resulting in a decrease of the specs loaded).
func SpecsLoaded(registryName string, specCount int) {
	defer recoverMetricPanic()
	specsLoaded.With(prometheus.Labels{"registry_name": registryName}).Add(float64(specCount))
}

// SpecsUnloaded - Will remove the count of specs. (The value can be negative,
// resulting in a increase in the number of specs loaded).
func SpecsUnloaded(registryName string, specCount int) {
	defer recoverMetricPanic()
	specsLoaded.With(prometheus.Labels{"registry_name": registryName}).Sub(float64(specCount))
}

// SpecsLoadedReset - Will reset all the values in in the gauge.
func SpecsLoadedReset() {
	defer recoverMetricPanic()
	specsLoaded.Reset()
}

// SpecsReset - Counter for how many times the specs are reloaded.
func SpecsReset() {
	defer recoverMetricPanic()
	specsReset.Inc()
}

// ProvisionJobStarted - Add a provision job to the counter.
func ProvisionJobStarted() {
	defer recoverMetricPanic()
	provisionJob.Inc()
}

// DeprovisionJobStarted - Add a deprovision job to the counter.
func DeprovisionJobStarted() {
	defer recoverMetricPanic()
	deprovisionJob.Inc()
}

// BindingJobStarted - Add a provision job to the counter.
func BindingJobStarted() {
	defer recoverMetricPanic()
	bindingJob.Inc()
}

// ProvisionJobFinished - Remove a provision job from the counter.
func ProvisionJobFinished() {
	defer recoverMetricPanic()
	provisionJob.Dec()
}

// DeprovisionJobFinished - Remove a deprovision job from the counter.
func DeprovisionJobFinished() {
	defer recoverMetricPanic()
	deprovisionJob.Dec()
}

// UpdateJobStarted - Add an update job to the counter.
func UpdateJobStarted() {
	defer recoverMetricPanic()
	updateJob.Inc()
}

// UpdateJobFinished - Remove an update job from the counter.
func UpdateJobFinished() {
	defer recoverMetricPanic()
	updateJob.Dec()
}

// BindingJobFinished - Remove a provision job from the counter.
func BindingJobFinished() {
	defer recoverMetricPanic()
	bindingJob.Dec()
}

// ActionStarted - Registers that an action has been started.
func ActionStarted(action string) {
	defer recoverMetricPanic()
	requests.WithLabelValues(action).Inc()
}
