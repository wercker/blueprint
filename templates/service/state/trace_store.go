package state

import (
	"fmt"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"
)

type TraceStore struct {
	store     Store
	tracer    opentracing.Tracer
	component string
}

func NewTraceStore(store Store, tracer opentracing.Tracer) *TraceStore {
	component := strings.TrimPrefix(fmt.Sprintf("%T", store), "*")
	return &TraceStore{
		store:     store,
		tracer:    tracer,
		component: component,
	}
}

func (s *TraceStore) Healthy() error {
	return s.store.Healthy()
}

func (s *TraceStore) Close() error {
	return s.store.Close()
}

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

var _ Store = (*TraceStore)(nil)
