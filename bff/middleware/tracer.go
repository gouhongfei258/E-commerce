package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const BFFTracerName = "github.com/storm/myidea/bff"

// Trace returns a Gin middleware that creates a root span for every request
// and injects the trace context into the request's context.
// When the downstream gRPC client is built with OpenTelemetry instrumentation,
// the traceid propagates transparently via gRPC metadata.
func Trace() gin.HandlerFunc {
	tp := otel.GetTracerProvider()
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Extract any incoming trace context from headers.
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		tracer := tp.Tracer(BFFTracerName)
		ctx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath(),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Propagate the trace context forward.
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Request.Header))

		// Attach a per-request timeout context.
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
