package state

import "io"

// Store provides access to data that is required for blueprint.
type Store interface {
	io.Closer
}
