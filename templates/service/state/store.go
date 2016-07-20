package state

import "io"

// Store provides access to data that is required for auth.
type Store interface {
	io.Closer
}
