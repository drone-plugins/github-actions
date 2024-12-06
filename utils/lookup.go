// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/exp/slog"
)

// ParseLookup parses the step string and returns the
// associated repository and ref.
func ParseLookup(s string) (repo string, ref string, ok bool) {
	org, repo, _, ref, err := parseActionName(s)
	if err == nil {
		url := fmt.Sprintf("https://github.com/%s/%s", org, repo)
		slog.Debug(fmt.Sprintf("parsed repo: %s, ref: %s", url, ref))
		return url, ref, true
	}

	slog.Warn(fmt.Sprintf("failed to parse action name: %s with err: %v", s, err))
	if !strings.HasPrefix(s, "https://github.com") {
		s, _ = url.JoinPath("https://github.com", s)
	}

	slog.Debug("parsed repo", s)
	if parts := strings.SplitN(s, "@", 2); len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return s, "", true
}
