// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"os"
	"testing"
)

// SkipForFerretDBv1 skips the test if running against FerretDB v1.
func SkipForFerretDBv1(tb testing.TB) {
	tb.Helper()

	v := os.Getenv("WIRE_FERRETDBV1")

	if v == "1" || v == "true" || v == "yes" {
		tb.Skip("Not implemented for v1")
	}
}
