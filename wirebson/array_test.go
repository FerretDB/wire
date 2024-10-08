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

package wirebson

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray(t *testing.T) {
	t.Parallel()

	arr := MustArray("foo", "bar", "baz")

	t.Run("All", func(t *testing.T) {
		t.Parallel()

		expected := map[int]any{
			0: "foo",
			1: "bar",
			2: "baz",
		}
		assert.Equal(t, expected, maps.Collect(arr.All()))
	})

	t.Run("Values", func(t *testing.T) {
		t.Parallel()

		expected := []any{"foo", "bar", "baz"}
		assert.Equal(t, expected, slices.Collect(arr.Values()))
	})
}
