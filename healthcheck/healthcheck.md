# Healthcheck Library

## General description

This healthcheck package provides a continuous monitoring framework that actively polls service dependencies and exposes their health status via an HTTP endpoint with integrated Prometheus metrics.

At the center is a registry that stores named health checkers and runs them on a recurring interval. Each checker is a component-specific probe that returns `nil` when healthy, or an error when unhealthy.

The registry operates with two background goroutines:

1. **CheckInLoop**: Runs all registered checkers concurrently at the configured interval
2. **ReadResults**: Collects check results from channels and updates the current health state

When a health request arrives, the HTTP handler returns the most recent cached state without triggering new checks.

The HTTP handler returns:

- `200 OK` when overall status is `UP`
- `503 Service Unavailable` when overall status is `DOWN`

The response body is JSON with the overall state and detailed component breakdown.

## How to use it on a service

Typical integration steps:

1. Create a registry with a check interval, Prometheus registerer, and metrics namespace.
2. Register one checker per dependency (database, cache, external API, queue, etc).
3. Mount `registry.Handler()` on an endpoint such as `/healthz`.
4. Let your orchestrator (Kubernetes, load balancer, platform probes) consume this endpoint.

Example:
```Go
package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/juliokobayashi-ebury/lib-teste/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
)

type DB struct{}

func (ref DB) Check(ctx context.Context) error {
	// Replace with a real dependency check.
	return nil
}

type Integration struct{}

func (Integration) Check(ctx context.Context) error {
	// Replace with real upstream call validation.
	return errors.New("upstream timeout")
}

func main() {
	// Create registry with 10-second check interval
	registry := healthcheck.NewRegistry(
		10*time.Second,
		prometheus.DefaultRegisterer,
		"myservice",
	)

	// Register checkers
	registry.Register("database", DB{})
	registry.Register("integration", Integration{})

	mux := http.NewServeMux()
	mux.Handle("/healthz", registry.Handler())

	_ = http.ListenAndServe(":8080", mux)
}
```

Expected behavior:

- If all components are healthy, endpoint returns `200` and status `UP`.
- If any component fails or times out, endpoint returns `503` and status `DOWN`.
- Health checks run automatically every 10 seconds in the background.
- Prometheus metrics are updated with each check cycle.

## Prometheus Metrics

The registry automatically exposes two Prometheus metrics:

1. **`<namespace>_healthcheck_availability`** (Gauge): Overall system availability
   - `1.0` when all components are UP
   - `0.0` when any component is DOWN

2. **`<namespace>_healthcheck_component_availability`** (GaugeVec): Per-component availability
   - Label: `component` (the registered checker name)
   - `1.0` when the component is UP
   - `0.0` when the component is DOWN

## API Reference

### `NewRegistry(interval time.Duration, prometheusInstance prometheus.Registerer, metricsNamespace string) *Registry`

Creates a new health check registry and starts background monitoring loops.

- **interval**: How often to run all health checks (also used as timeout per check)
- **prometheusInstance**: Prometheus registerer for metrics (uses `prometheus.DefaultRegisterer` if `nil`)
- **metricsNamespace**: Namespace prefix for Prometheus metrics

### `Register(name string, checker Checker) error`

Registers a new health checker. Returns error if a checker with the same name already exists.

### `Unregister(name string) error`

Removes a health checker from the registry. Returns error if the checker doesn't exist.

### `Handler() http.HandlerFunc`

Returns an HTTP handler that serves the current health status as JSON.

## Principles

This implementation follows practical principles for production monitoring:

### 1. Continuous monitoring with cached results

Background checks run on an interval. HTTP requests return cached state instantly without triggering new checks, avoiding request pile-up during outages.

### 2. Fail fast and fail clearly

Each checker returns an explicit error. The response includes per-component errors with timestamps so operators can identify what failed and when.

### 3. Concurrency by default

Checks run in parallel to keep check cycle time close to the slowest check, not the sum of all checks.

### 4. Bounded execution time

Each checker receives a context with timeout equal to the check interval, preventing hangs and protecting the service under dependency slowness.

### 5. Aggregated but transparent status

Overall status is simple (`UP` or `DOWN`), while component-level details preserve observability. Any single component failure sets the overall status to `DOWN`.

### 6. Prometheus integration

Built-in Prometheus metrics allow monitoring systems to track availability over time, alert on degradation, and calculate SLIs.
