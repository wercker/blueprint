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
	"fmt"
	"strconv"
	"time"
)

var (
	// GitCommit is the git commit hash associated with this build.
	GitCommit = ""

	// MajorVersion is the semver major version.
	MajorVersion = "1"

	// MinorVersion is the semver minor version.
	MinorVersion = "0"

	// PatchVersion is the semver patch version. (use 0 for dev, build process
	// will inject a build number)
	PatchVersion = "0"

	// Compiled is the unix timestamp when this binary got compiled.
	Compiled = ""
)

func init() {
	if Compiled == "" {
		Compiled = strconv.FormatInt(time.Now().Unix(), 10)
	}
}

// Version returns a semver compatible version for this build.
func Version() string {
	return fmt.Sprintf("%s.%s.%s", MajorVersion, MinorVersion, PatchVersion)
}

// CompiledAt converts the Unix time Compiled to a time.Time using UTC timezone.
func CompiledAt() time.Time {
	i, err := strconv.ParseInt(Compiled, 10, 64)
	if err != nil {
		panic(err)
	}

	return time.Unix(i, 0).UTC()
}
