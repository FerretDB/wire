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
	"encoding/binary"
	"log/slog"
	"strconv"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
	"github.com/FerretDB/wire/internal/util/must"
)

// Array represents a BSON array in the (partially) decoded form.
type Array struct {
	elements []any
	frozen   bool
}

// NewArray creates a new Array from the given values.
func NewArray(values ...any) (*Array, error) {
	res := &Array{
		elements: make([]any, 0, len(values)),
	}

	for i, v := range values {
		if err := res.Add(v); err != nil {
			return nil, lazyerrors.Errorf("%d: %w", i, err)
		}
	}

	return res, nil
}

// MustArray is a variant of [NewArray] that panics on error.
func MustArray(values ...any) *Array {
	res, err := NewArray(values...)
	if err != nil {
		panic(err)
	}

	return res
}

// MakeArray creates a new empty Array with the given capacity.
func MakeArray(cap int) *Array {
	return &Array{
		elements: make([]any, 0, cap),
	}
}

// Freeze prevents array from further modifications.
// Any methods that would modify the array will panic.
//
// It is safe to call Freeze multiple times.
func (arr *Array) Freeze() {
	arr.frozen = true
}

// checkFrozen panics if array is frozen.
func (arr *Array) checkFrozen() {
	if arr.frozen {
		panic("array is frozen and can't be modified")
	}
}

// Len returns the number of elements in the Array.
func (arr *Array) Len() int {
	return len(arr.elements)
}

// Get returns the element at the given index.
// It panics if index is out of bounds.
func (arr *Array) Get(index int) any {
	return arr.elements[index]
}

// Add adds a new element to the end of the Array.
func (arr *Array) Add(value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Error(err)
	}

	arr.checkFrozen()

	arr.elements = append(arr.elements, value)

	return nil
}

// Replace sets the value of the element at the given index.
// It panics if index is out of bounds.
func (arr *Array) Replace(index int, value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Error(err)
	}

	arr.checkFrozen()

	arr.elements[index] = value

	return nil
}

// Encode encodes non-nil BSON array into raw.
//
// The function operates directly on raw RawArray.
// It doesn't reallocate memory, hence raw needs to have the proper length.
func (arr *Array) Encode(raw RawArray) error {
	must.NotBeZero(arr)

	binary.LittleEndian.PutUint32(raw[0:4], uint32(sizeArray(arr)))

	index := 4
	for i, v := range arr.elements {
		written, err := encodeField(raw[index:], strconv.Itoa(i), v)
		if err != nil {
			return lazyerrors.Error(err)
		}

		index += written
	}

	raw[index] = byte(0)
	index++

	return nil
}

// Decode returns itself to implement [AnyArray].
//
// Receiver must not be nil.
func (arr *Array) Decode() (*Array, error) {
	must.NotBeZero(arr)
	return arr, nil
}

// LogValue implements [slog.LogValuer].
func (arr *Array) LogValue() slog.Value {
	return slogValue(arr, 1)
}

// check interfaces
var (
	_ AnyArray       = (*Array)(nil)
	_ slog.LogValuer = (*Array)(nil)
)
