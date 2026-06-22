package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

type Checker interface {
	Check(ctx context.Context) error
}

type CheckerFunc func(ctx context.Context) error

func (f CheckerFunc) Check(ctx context.Context) error {
	return f(ctx)
}

type ComponentResult struct {
	Status    Status    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type OverallResponse struct {
	Status               Status                     `json:"status"`
	CapabilityPercentage float64                    `json:"capability_percentage"`
	Components           map[string]ComponentResult `json:"components"`
}

type Registry struct {
	mu       sync.RWMutex
	checkers map[string]Checker
	timeout  time.Duration
}

func NewRegistry(timeout time.Duration) *Registry {
	return &Registry{
		checkers: make(map[string]Checker),
		timeout:  timeout,
	}
}

func (r *Registry) Register(name string, checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[name] = checker
}

func (r *Registry) Check(ctx context.Context) OverallResponse {
	r.mu.RLock()
	defer r.mu.RUnlock()

	checkCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex

	response := OverallResponse{
		Status:     StatusUp,
		Components: make(map[string]ComponentResult),
	}

	for name, checker := range r.checkers {
		wg.Add(1)
		go func(name string, chk Checker) {
			defer wg.Done()

			errCh := make(chan error, 1)
			go func() {
				errCh <- chk.Check(checkCtx)
			}()

			var err error
			select {
			case e := <-errCh:
				err = e
			case <-checkCtx.Done():
				err = checkCtx.Err()
			}

			status := StatusUp
			errMsg := ""
			if err != nil {
				status = StatusDown
				errMsg = err.Error()
			}

			mu.Lock()
			response.Components[name] = ComponentResult{
				Status:    status,
				Error:     errMsg,
				Timestamp: time.Now().UTC(),
			}
			if status == StatusDown {
				response.Status = StatusDown
			}
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()

	totalComponents := len(response.Components)
	healthyComponents := 0

	for _, comp := range response.Components {
		if comp.Status == StatusUp {
			healthyComponents++
		}
	}

	if totalComponents > 0 {
		response.CapabilityPercentage = (float64(healthyComponents) / float64(totalComponents)) * 100.0
	} else {
		response.CapabilityPercentage = 100.0
	}

	return response
}

func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		response := r.Check(req.Context())

		w.Header().Set("Content-Type", "application/json")

		if response.Status == StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(response)
	}
}

type HealthMetrics interface {
	RegisterAvailability(status Status, availability float64, timestamp time.Time)
}
