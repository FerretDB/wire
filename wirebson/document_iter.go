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
	"iter"
)

// All returns a Seq2 that yields all field name value pairs of the document.
func (doc *Document) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for _, f := range doc.fields {
			if !yield(f.name, f.value) {
				return
			}
		}
	}
}
