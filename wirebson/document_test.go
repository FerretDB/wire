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
	"time"

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
	original := MustDocument(
		"doc", MustDocument("key1", "value1"),
		"array", MustArray("v1", "v2"),
		"double", float64(0),
		"string", "foo",
		"undefined", Undefined,
		"objectID", ObjectID{},
		"boolean", false,
		"time", time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		"null", Null,
		"regex", Regex{Pattern: "foo", Options: "bar"},
		"int", int32(0),
		"timestamp", Timestamp(0),
		"long", int64(0),
		"decimal", Decimal128{L: 0, H: 0},
		"binary", Binary{B: []byte{0, 0, 0, 0, 0, 0}, Subtype: BinaryGeneric},
	)

	cp := original.Copy()
	require.Equal(t, original, cp)
	require.NotSame(t, original, cp)

	originalBinary := original.Get("binary").(Binary).B
	copyBinary := cp.Get("binary").(Binary).B
	require.Equal(t, originalBinary, copyBinary)
	require.NotSame(t, &originalBinary[0], &copyBinary[0])
}
