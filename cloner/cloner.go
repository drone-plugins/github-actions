// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cloner provides support for cloning git repositories.
package cloner

import (
	"context"
)

type (
	// Params provides clone params.
	Params struct {
		Repo string
		Ref  string
		Sha  string
		Dir  string // Target clone directory.
	}

	// Cloner clones a repository.
	Cloner interface {
		// Clone a repository.
		Clone(context.Context, Params) error
	}
)
