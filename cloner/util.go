// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloner

import (
	"regexp"
	"strings"
)

// regular expressions to test whether or not a string is
// a sha1 or sha256 commit hash.
var (
	sha1   = regexp.MustCompile("^([a-f0-9]{40})$")
	sha256 = regexp.MustCompile("^([a-f0-9]{64})$")
	semver = regexp.MustCompile(`^v?((([0-9]+)(?:\.([0-9]+))?(?:\.([0-9]+))?(?:-([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?)(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?)$`)
)

// helper function returns true if the string is a commit hash.
func isHash(s string) bool {
	return sha1.MatchString(s) || sha256.MatchString(s)
}

// helper function returns the branch name expanded to the
// fully qualified reference path (e.g refs/heads/master).
func expandRef(name string) string {
	if strings.HasPrefix(name, "refs/") {
		return name
	}
	if semver.MatchString(name) {
		return "refs/tags/" + name
	}
	return "refs/heads/" + name
}
