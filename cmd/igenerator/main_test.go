//-----------------------------------------------------------------------------
// Copyright (c) 2017 Oracle and/or its affiliates.  All rights reserved.
// This program is free software: you can modify it and/or redistribute it
// under the terms of:
//
// (i)  the Universal Permissive License v 1.0 or at your option, any
//      later version (http://oss.oracle.com/licenses/upl); and/or
//
// (ii) the Apache License v 2.0. (http://www.apache.org/licenses/LICENSE-2.0)
//-----------------------------------------------------------------------------

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_contains(t *testing.T) {
	tests := []struct {
		needle   string
		haystack []string
		expected bool
	}{
		{"foofoo", []string{}, false},
		{"foofoo", []string{"foofoo"}, true},
		{"foofoo", []string{"barbar", "foofoo"}, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			actual := contains(tt.needle, tt.haystack)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
