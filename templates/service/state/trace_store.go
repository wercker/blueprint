package state

import (
	"fmt"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"
)

// NewTraceStore creates a new TraceStore.
func NewTraceStore(store Store, tracer opentracing.Tracer) *TraceStore {
	component := strings.TrimPrefix(fmt.Sprintf("%T", store), "*")
	return &TraceStore{
		store:     store,
		tracer:    tracer,
		component: component,
	}
}

// TraceStore wraps another Store and sends trace information to zipkin.
type TraceStore struct {
	store     Store
	tracer    opentracing.Tracer
	component string
}

var _ Store = (*TraceStore)(nil)

// TODO: add methods here

func (s *TraceStore) trace(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (context.Context, opentracing.Span) {
	var span opentracing.Span
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
		span = s.tracer.StartSpan(operationName, opts...)
	} else {
		span = s.tracer.StartSpan(operationName, opts...)
	}

	span.SetTag(string(ext.Component), s.component)

	return opentracing.ContextWithSpan(ctx, span), span
}

// Healthy calls Healthy on the wrapped store.
func (s *TraceStore) Healthy() error {
	return s.store.Healthy()
}

// Close calls Close on the wrapped store.
func (s *TraceStore) Close() error {
	return s.store.Close()
}
