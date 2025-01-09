// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloner

import (
	"context"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	maxRetries      = 3
	backoffInterval = time.Second * 1
)

// New returns a new cloner.
func New(depth int, stdout io.Writer) Cloner {
	c := &cloner{
		depth:  depth,
		stdout: stdout,
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		c.username = "token"
		c.password = token
	}
	return c
}

// NewDefault returns a cloner with default settings.
func NewDefault() Cloner {
	return New(1, os.Stdout)
}

// default cloner using the built-in Git client.
type cloner struct {
	depth    int
	username string
	password string
	stdout   io.Writer
}

// Clone the repository using the built-in Git client.
func (c *cloner) Clone(ctx context.Context, params Params) error {
	opts := &git.CloneOptions{
		RemoteName: "origin",
		Progress:   c.stdout,
		URL:        params.Repo,
		Tags:       git.NoTags,
	}
	// set the reference name if provided
	if params.Ref != "" {
		opts.ReferenceName = plumbing.ReferenceName(expandRef(params.Ref))
	}
	// set depth if cloning the head commit of a branch as
	// opposed to a specific commit sha
	if params.Sha == "" {
		opts.Depth = c.depth
	}
	if c.username != "" && c.password != "" {
		opts.Auth = &http.BasicAuth{
			Username: c.username,
			Password: c.password,
		}
	}
	// clone the repository
	var (
		r   *git.Repository
		err error
	)

	retryStrategy := backoff.NewExponentialBackOff()
	retryStrategy.InitialInterval = backoffInterval
	retryStrategy.MaxInterval = backoffInterval * 5     // Maximum delay
	retryStrategy.MaxElapsedTime = backoffInterval * 60 // Maximum time to retry (1min)

	b := backoff.WithMaxRetries(retryStrategy, uint64(maxRetries))

	err = backoff.Retry(func() error {
		r, err = git.PlainClone(params.Dir, false, opts)
		if err == nil {
			return nil
		}
		if (errors.Is(plumbing.ErrReferenceNotFound, err) || matchRefNotFoundErr(err)) &&
			!strings.HasPrefix(params.Ref, "refs/") {
			originalRefName := opts.ReferenceName
			// If params.Ref is provided without refs/*, then we are assuming it to either refs/heads/ or refs/tags.
			// Try clone again with inverse ref.
			if opts.ReferenceName.IsBranch() {
				opts.ReferenceName = plumbing.ReferenceName("refs/tags/" + params.Ref)
			} else if opts.ReferenceName.IsTag() {
				opts.ReferenceName = plumbing.ReferenceName("refs/heads/" + params.Ref)
			} else {
				return err // Return err if the reference name is invalid
			}

			r, err = git.PlainClone(params.Dir, false, opts)
			if err == nil {
				return nil
			}
			// Change reference name back to original
			opts.ReferenceName = originalRefName
		}
		return err
	}, b)

	// If error not nil, then return it
	if err != nil {
		return err
	}

	if params.Sha == "" {
		return nil
	}

	// checkout the sha
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	return w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(params.Sha),
	})
}

func matchRefNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	pattern := `couldn't find remote ref.*`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(err.Error())
}
