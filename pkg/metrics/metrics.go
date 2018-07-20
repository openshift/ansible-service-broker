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
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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

	specsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "specs_total",
			Help:      "Spec count of different registries and marked for deletion.",
		}, []string{
			"source",
		},
	)

	specsDeleted = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: subsystem,
		Name:      "specs_deleted",
		Help:      "Specs deleted from data-store.",
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

	unbindingJob = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "unbinding_jobs",
			Help:      "How many unbinding jobs are actively in the buffer.",
		})

	requests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "actions_requested",
			Help:      "How many actions have been made.",
		}, []string{"action"})
)

func init() {
	prometheus.MustRegister(sandbox)
	prometheus.MustRegister(specsTotal)
	prometheus.MustRegister(specsDeleted)
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

// SpecsLoaded - Will add the count of specs.
func SpecsLoaded(registryName string, specCount int) {
	defer recoverMetricPanic()
	specsTotal.With(map[string]string{"source": registryName}).Set(float64(specCount))
}

// SpecsMarkedForDeletion - will add the number of specs marked for deletion
func SpecsMarkedForDeletion(specCount int) {
	defer recoverMetricPanic()
	specsTotal.With(map[string]string{"source": "marked_for_deletion"}).Set(float64(specCount))
}

// SpecsDeleted - will add the number of specs deleted from the data-store
func SpecsDeleted(specCount int) {
	defer recoverMetricPanic()
	specsDeleted.Add(float64(specCount))
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

// BindJobStarted - Add a provision job to the counter.
func BindJobStarted() {
	defer recoverMetricPanic()
	bindingJob.Inc()
}

// UnbindJobStarted - Add a provision job to the counter.
func UnbindJobStarted() {
	defer recoverMetricPanic()
	unbindingJob.Inc()
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

// BindJobFinished - Remove a provision job from the counter.
func BindJobFinished() {
	defer recoverMetricPanic()
	bindingJob.Dec()
}

// UnbindJobFinished - Remove a provision job from the counter.
func UnbindJobFinished() {
	defer recoverMetricPanic()
	unbindingJob.Dec()
}

// ActionStarted - Registers that an action has been started.
func ActionStarted(action string) {
	defer recoverMetricPanic()
	requests.WithLabelValues(action).Inc()
}
