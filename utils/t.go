// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

// spec defines the bitrise plugin.
type spec struct {
	Outputs map[string]interface{} `yaml:"outputs"`
}