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
	"sort"
)

// arraySort implements [sort.Interface].
type arraySort struct {
	arr  *Array
	less func(a, b any) bool
}

// Len implements [sort.Interface].
func (as arraySort) Len() int {
	return as.arr.Len()
}

// Less implements [sort.Interface].
func (as arraySort) Less(i int, j int) bool {
	return as.less(as.arr.Get(i), as.arr.Get(j))
}

// Swap implements [sort.Interface].
func (as arraySort) Swap(i int, j int) {
	a := as.arr.Get(i)
	as.arr.Replace(i, as.arr.Get(j))
	as.arr.Replace(j, a)
}

// check interfaces
var (
	_ sort.Interface = arraySort{}
)
