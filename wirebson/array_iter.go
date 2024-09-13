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
	"iter"
)

// All returns an iterator over index value pairs of the array.
func (arr *Array) All() iter.Seq2[int, any] {
	return func(yield func(int, any) bool) {
		for i, v := range arr.elements {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Values returns an iterator over values of the array.
func (arr *Array) Values() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range arr.elements {
			if !yield(v) {
				return
			}
		}
	}
}
