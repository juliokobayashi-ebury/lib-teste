package tracing

import (
	"github.com/labstack/echo/v4"
)

const (
	TraceIdHeader = "X-Request-Id"
)

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			
			traceId := r.Header.Get(TraceIdHeader)
			ctx := NewContextWithTracing(r.Context(), traceId)

			c.SetRequest(r.WithContext(ctx))
			return next(c)
		}
	}
}
