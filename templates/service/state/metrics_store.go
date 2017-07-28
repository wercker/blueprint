package state

import (
	"github.com/wercker/pkg/metrics"
)

// NewMetricsStore creates a new MetricsStore.
func NewMetricsStore(wrappedStore Store) *MetricsStore {
	store := &MetricsStore{store: wrappedStore, observer: metrics.NewStoreObserver()}
	store.observer.Preload(store)

	return store
}

// MetricsStore wraps another Store and sends metrics to Prometheus.
type MetricsStore struct {
	store    Store
	observer *metrics.StoreObserver
}

var _ Store = (*MetricsStore)(nil)

// TODO: Add methods here

// Close calls Close on the wrapped store.
func (s *MetricsStore) Close() error {
	return s.store.Close()
}

// Healthy calls Healthy on the wrapped store.
func (s *MetricsStore) Healthy() error {
	return s.store.Healthy()
}
