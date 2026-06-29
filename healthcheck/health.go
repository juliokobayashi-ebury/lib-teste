package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"

	MetricName = "healthcheck_availability"
	MetricHelp = "Availability of the system based on health checks"
)

var availabilityGaugeVec *prometheus.GaugeVec

type Checker interface {
	Check(ctx context.Context) error
}

type ComponentResult struct {
	Status    Status    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type OverallResponse struct {
	Status     Status                     `json:"status"`
	Components map[string]ComponentResult `json:"components"`
}

type Registry struct {
	checkersMutex   sync.RWMutex
	checkers        map[string]Checker
	checkerChannels map[string]chan error
	interval        time.Duration

	healthMutex   sync.RWMutex
	currentHealth OverallResponse
	// namespace string

	overallGauge   prometheus.Gauge
	componentGauge *prometheus.GaugeVec
}

func NewRegistry(interval time.Duration, prometheusInstance prometheus.Registerer, metricsNamespace string) *Registry {
	if prometheusInstance == nil {
		if prometheus.DefaultRegisterer == nil {
			prometheusInstance = prometheus.NewRegistry()
			prometheus.DefaultRegisterer = prometheusInstance
		} else {
			prometheusInstance = prometheus.DefaultRegisterer
		}
	}

	overallGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      MetricName,
		Help:      MetricHelp,
	})
	prometheusInstance.MustRegister(overallGauge)

	componentGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "healthcheck_component_availability",
		Help:      "Availability of individual components based on health checks",
	}, []string{"component"})
	prometheusInstance.MustRegister(componentGauge)

	registry := &Registry{
		checkers: make(map[string]Checker),
		interval: interval,

		currentHealth: OverallResponse{
			Status:     StatusUp,
			Components: make(map[string]ComponentResult),
		},
		checkerChannels: make(map[string]chan error),
		overallGauge:    overallGauge,
		componentGauge:  componentGauge,
	}

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go registry.CheckInLoop(interval)
	go registry.ReadResults(quitChannel)

	return registry
}

func (r *Registry) Register(name string, checker Checker) error {
	r.checkersMutex.Lock()
	r.healthMutex.Lock()
	defer r.checkersMutex.Unlock()
	defer r.healthMutex.Unlock()
	if _, exists := r.checkers[name]; exists {
		return fmt.Errorf("checker with name %s already exists", name)
	}

	r.checkers[name] = checker
	r.checkerChannels[name] = make(chan error)
	r.currentHealth.Components[name] = ComponentResult{
		Status:    StatusUp,
		Timestamp: time.Now(),
	}

	return nil
}

func (r *Registry) Unregister(name string) error {
	r.checkersMutex.Lock()
	r.healthMutex.Lock()
	defer r.healthMutex.Unlock()
	defer r.checkersMutex.Unlock()
	if _, exists := r.checkers[name]; !exists {
		return fmt.Errorf("checker with name %s does not exist", name)
	}
	delete(r.checkerChannels, name)
	delete(r.checkers, name)
	delete(r.currentHealth.Components, name)

	return nil
}

func (r *Registry) CheckInLoop(duration time.Duration) {
	defer time.AfterFunc(duration, func() {
		r.CheckInLoop(duration)
	})

	r.checkersMutex.RLock()
	defer r.checkersMutex.RUnlock()

	for name, checker := range r.checkers {
		go func(name string, cker Checker) {
			checkCtx, cancel := context.WithTimeout(context.Background(), r.interval)
			defer cancel()
			r.checkerChannels[name] <- cker.Check(checkCtx)
		}(name, checker)
	}
}

func (r *Registry) ReadResults(quitChannel chan os.Signal) {
	defer time.AfterFunc(r.interval, func() {
		r.ReadResults(quitChannel)
	})

	var wg sync.WaitGroup

	for name, ch := range r.checkerChannels {
		wg.Add(1)
		go func(name string, ch chan error) {
			defer wg.Done()
			select {
			case err := <-ch:
				r.healthMutex.Lock()
				componentStatus := 1.0
				if err != nil {
					componentStatus = 0.0
					r.currentHealth.Components[name] = ComponentResult{
						Status:    StatusDown,
						Error:     err.Error(),
						Timestamp: time.Now(),
					}
				} else {
					r.currentHealth.Components[name] = ComponentResult{
						Status:    StatusUp,
						Error:     "",
						Timestamp: time.Now(),
					}
				}
				r.componentGauge.WithLabelValues(name).Set(componentStatus)
				r.healthMutex.Unlock()
			case <-quitChannel:
				return
			}
		}(name, ch)
	}
	wg.Wait()

	r.UpdateResponse()
}

func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		r.healthMutex.RLock()
		response := r.currentHealth
		r.healthMutex.RUnlock()

		if response.Status == StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(response)
	}
}

func (r *Registry) UpdateResponse() {
	r.healthMutex.Lock()
	defer r.healthMutex.Unlock()

	overallStatus := 1.0

	for _, component := range r.currentHealth.Components {
		if component.Status != StatusUp {
			r.currentHealth.Status = StatusDown

			overallStatus = 0.0
			break
		}
	}
	if overallStatus == 1.0 {
		r.currentHealth.Status = StatusUp
	}
	r.overallGauge.Set(overallStatus)
}
