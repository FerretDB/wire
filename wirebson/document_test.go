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
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocument(t *testing.T) {
	t.Parallel()

	doc := MustDocument(
		"foo", int32(1),
		"bar", int32(2),
		"baz", int64(3),
	)

	t.Run("All", func(t *testing.T) {
		t.Parallel()

		expected := map[string]any{
			"foo": int32(1),
			"bar": int32(2),
			"baz": int64(3),
		}

		assert.Equal(t, expected, maps.Collect(doc.All()))
	})

	t.Run("Fields", func(t *testing.T) {
		t.Parallel()

		expected := []string{"foo", "bar", "baz"}
		assert.Equal(t, expected, slices.Collect(doc.Fields()))
	})
}

func TestDocumentCopy(t *testing.T) {
	o := MustDocument(
		"nan", math.NaN(),
		"binary", Binary{B: []byte{0, 1, 2, 3, 4, 5}, Subtype: BinaryVector},
	)

	c := o.Copy()
	assertEqual(t, o, c)
	require.NotSame(t, o, c)

	ob := o.Get("binary").(Binary).B
	cb := c.Get("binary").(Binary).B
	require.Equal(t, ob, cb)
	require.NotSame(t, unsafe.SliceData(ob), unsafe.SliceData(cb))
}
