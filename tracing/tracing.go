package tracing

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strings"
)

const (
	TraceIdField = "trace_id"
)

func NewContextWithTracing(ctx context.Context, optionalTraceId ...string) context.Context {
	traceId := ""
	if len(optionalTraceId) > 0 {
		traceId = strings.TrimSpace(optionalTraceId[0])
	}

	if traceId == "" {
		traceId = uuid.NewString()
	}

	ctx = context.WithValue(ctx, TraceIdField, traceId)
	ctx = log.Ctx(ctx).
		With().
		Str(TraceIdField, traceId).
		Logger().
		WithContext(ctx)

	return ctx
}

func TraceIdFromContext(ctx context.Context) string {
	traceId, _ := ctx.Value(TraceIdField).(string)
	return traceId
}
