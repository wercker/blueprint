package state

import (
	_ "context"

	"github.com/wercker/pkg/metrics"
)

// MetricsStore wraps another Store and sends metrics to Prometheus.
type MetricsStore struct {
	store    Store
	observer *metrics.StoreObserver
}

// NewMetricsStore creates a new MetricsStore.
func NewMetricsStore(wrappedStore Store) *MetricsStore {
	store := &MetricsStore{store: wrappedStore, observer: metrics.NewStoreObserver()}

	return store
}

var _ Store = (*MetricsStore)(nil)

{{range .}}
{{template "doc" . -}}
func (s *MetricsStore) {{.Name}}({{template "list" .Params}}) ({{template "list" .Returns}}) {
	done := s.observer.Observe("{{.Name}}")
	defer done()
	
	return s.store.{{.Name}}({{template "call" .Params}})
}
{{end}}

// Initialize calls Initialize on the wrapped store.
func (s *MetricsStore) Initialize() error {
	s.observer.Preload(s, "Initialize")
	return s.store.Initialize()
}

// Healthy calls Healthy on the wrapped store.
func (s *MetricsStore) Healthy() error {
	return s.store.Healthy()
}

// Close calls Close on the wrapped store.
func (s *MetricsStore) Close() error {
	return s.store.Close()
}

{{define "list"}}{{range $index, $element := .}}{{if $index}}, {{end}}{{if $element.Name}}{{$element.Name}}{{end}} {{$element.Type}}{{end}}{{end}}
{{define "call"}}{{range $index, $element := .}}{{if $index}}, {{end}}{{if $element.Name}}{{$element.Name}}{{end}}{{end}}{{end}}
{{define "doc"}}
{{range .Doc}}
{{.}}
{{- else}}
// {{.Name}} .
{{- end}}
{{end}}
