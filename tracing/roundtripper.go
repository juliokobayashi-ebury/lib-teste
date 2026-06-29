package tracing

import (
	"net/http"
)

type tracingTransport struct {
	next http.RoundTripper
}

func NewTransport(next http.RoundTripper) http.RoundTripper {
	return &tracingTransport{next: next}
}

func (ref *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	traceId := TraceIdFromContext(req.Context())
	if traceId != "" {
		req.Header.Set(TraceIdHeader, traceId)
	}

	return ref.next.RoundTrip(req)
}
