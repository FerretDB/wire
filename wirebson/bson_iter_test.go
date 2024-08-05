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

// We should remove the build tag below once we use 1.23 everywhere.
// See https://pkg.go.dev/internal/goexperiment
// TODO https://github.com/FerretDB/wire/issues/9

//go:build goexperiment.rangefunc

package wirebson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentAll(t *testing.T) {
	t.Parallel()

	doc := MustDocument(
		"foo", int32(1),
		"bar", int32(2),
		"baz", int64(3),
	)

	var ks []string
	var vs []any

	for k, v := range doc.All() {
		ks = append(ks, k)
		vs = append(vs, v)
	}

	require.Equal(t, []string{"foo", "bar", "baz"}, ks)
	require.Equal(t, []any{int32(1), int32(2), int64(3)}, vs)
}

func TestArrayAll(t *testing.T) {
	t.Parallel()

	arr := MustArray("foo", int32(1), "bar", int32(2))

	var vs []any
	for v := range arr.All() {
		vs = append(vs, v)
	}

	require.Equal(t, []any{"foo", int32(1), "bar", int32(2)}, vs)
}
