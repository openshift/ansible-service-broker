package metrics

import (
	logging "github.com/op/go-logging"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	subsystem = "asb"
)

var (
	sandboxCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "sandbox_created",
			Help:      "Counter of all sandbox namespaces that are created.",
		})

	sandboxDeleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "sandbox_deleted",
			Help:      "Counter of all sandbox namespaces that are deleted.",
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

	log = logging.MustGetLogger("metrics")
)

func init() {
	prometheus.MustRegister(sandboxCreated)
	prometheus.MustRegister(sandboxDeleted)
	prometheus.MustRegister(specsLoaded)
	prometheus.MustRegister(specsReset)
	prometheus.MustRegister(provisionJob)
	prometheus.MustRegister(deprovisionJob)
}

// Init - Initialize the metrics package.
func Init(logger *logging.Logger) {
	log = logger
}

func recoverMetricPanic() {
	if r := recover(); r != nil {
		log.Errorf("Recovering from metric function - %v", r)
	}
}

// SandboxCreated - Counter for how many sandbox created.
func SandboxCreated() {
	defer recoverMetricPanic()
	sandboxCreated.Inc()
}

// SandboxDeleted - Counter for how many sandbox deleted.
func SandboxDeleted() {
	defer recoverMetricPanic()
	sandboxDeleted.Inc()
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

// AddProvisionJob - Add a provision job to the counter.
func AddProvisionJob() {
	defer recoverMetricPanic()
	provisionJob.Inc()
}

// AddDeprovisionJob - Add a deprovision job to the counter.
func AddDeprovisionJob() {
	defer recoverMetricPanic()
	deprovisionJob.Inc()
}

// RemoveProvisionJob - Remove a provision job to the counter.
func RemoveProvisionJob() {
	defer recoverMetricPanic()
	provisionJob.Dec()
}

// RemoveDeprovisionJob - Remove a deprovision job to the counter.
func RemoveDeprovisionJob() {
	defer recoverMetricPanic()
	deprovisionJob.Dec()
}
