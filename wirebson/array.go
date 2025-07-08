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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"iter"
	"log/slog"
	"slices"
	"sort"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/FerretDB/wire/internal/util/lazyerrors"
	"github.com/FerretDB/wire/internal/util/must"
)

// Array represents a BSON array in the (partially) decoded form.
type Array struct {
	values []any
	frozen bool
}

// NewArray creates a new Array from the given values.
func NewArray(values ...any) (*Array, error) {
	res := &Array{
		values: make([]any, 0, len(values)),
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
		values: make([]any, 0, cap),
	}
}

func (arr *Array) array() {}

// Freeze prevents Array from further modifications.
// Any methods that would modify the Array will panic.
//
// It is safe to call Freeze multiple times.
func (arr *Array) Freeze() {
	arr.frozen = true
}

// checkFrozen panics if Array is frozen.
func (arr *Array) checkFrozen() {
	if arr.frozen {
		panic("array is frozen and can't be modified")
	}
}

// Len returns the number of values in the Array.
func (arr *Array) Len() int {
	return len(arr.values)
}

// Get returns the value at the given index.
// It panics if index is out of bounds.
func (arr *Array) Get(index int) any {
	return arr.values[index]
}

// All returns an iterator over index/value pairs of the Array.
func (arr *Array) All() iter.Seq2[int, any] {
	return func(yield func(int, any) bool) {
		for i, v := range arr.values {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Values returns an iterator over values of the Array.
func (arr *Array) Values() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range arr.values {
			if !yield(v) {
				return
			}
		}
	}
}

// Add adds a new value to the end of the Array.
func (arr *Array) Add(value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Error(err)
	}

	arr.checkFrozen()

	arr.values = append(arr.values, value)

	return nil
}

// Replace sets the value of the element at the given index.
// It panics if index is out of bounds.
func (arr *Array) Replace(index int, value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Error(err)
	}

	arr.checkFrozen()

	arr.values[index] = value

	return nil
}

// SortInterface returns [sort.Interface] that can be used to sort Array in place.
// Passed function should return true is a < b, false otherwise.
// It should be able to handle values of different types.
func (arr *Array) SortInterface(less func(a, b any) bool) sort.Interface {
	return arraySort{
		arr:  arr,
		less: less,
	}
}

// Encode encodes non-nil Array.
//
// TODO https://github.com/FerretDB/wire/issues/21
// This method should accept a slice of bytes, not return it.
// That would allow to avoid unnecessary allocations.
func (arr *Array) Encode() (RawArray, error) {
	must.NotBeZero(arr)

	size := sizeArray(arr)
	buf := bytes.NewBuffer(make([]byte, 0, size))

	if err := binary.Write(buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for i, v := range arr.values {
		if err := encodeField(buf, strconv.Itoa(i), v); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, byte(0)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

// MarshalJSON implements [json.Marshaler]
// by encoding Canonical Extended JSON v2 representation of the array.
func (arr *Array) MarshalJSON() ([]byte, error) {
	// encoding/json does not call this method on nil
	must.NotBeZero(arr)

	a, err := ToDriver(arr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	b, err := bson.MarshalExtJSON(a, true, false)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}

// Decode returns itself to implement [AnyArray].
//
// Receiver must not be nil.
func (arr *Array) Decode() (*Array, error) {
	must.NotBeZero(arr)
	return arr, nil
}

// UnmarshalJSON implements [json.Unmarshaler]
// by decoding Canonical Extended JSON v2 representation of the array.
func (arr *Array) UnmarshalJSON(b []byte) error {
	// encoding/json does not call this method on nil
	must.NotBeZero(arr)

	var a bson.A
	if err := bson.UnmarshalExtJSON(b, true, &a); err != nil {
		return lazyerrors.Error(err)
	}

	v, err := FromDriver(a)
	if err != nil {
		return lazyerrors.Error(err)
	}

	switch v := v.(type) {
	case *Array:
		must.NotBeZero(v)
		*arr = *v
		return nil
	default:
		return lazyerrors.Errorf("expected *Array, got %T", v)
	}
}

// Copy returns a shallow copy of [*Array]. Only scalar values (including [Binary]) are copied.
// [*Document], [*Array], [RawDocument], and [RawArray] are added without a copy, using the same pointer/slice.
func (arr *Array) Copy() *Array {
	res := MakeArray(arr.Len())

	for v := range arr.Values() {
		switch v := v.(type) {
		case Binary:
			must.NoError(res.Add(Binary{B: slices.Clip(slices.Clone(v.B)), Subtype: v.Subtype}))
		default:
			must.NoError(validBSONType(v))
			must.NoError(res.Add(v))
		}
	}

	return res
}

// LogValue implements [slog.LogValuer].
func (arr *Array) LogValue() slog.Value {
	return slogValue(arr, 1)
}

// LogMessage implements [AnyArray].
func (arr *Array) LogMessage() string {
	return LogMessage(arr)
}

// LogMessageIndent implements [AnyArray].
func (arr *Array) LogMessageIndent() string {
	return LogMessageIndent(arr)
}

// check interfaces
var (
	_ AnyArray         = (*Array)(nil)
	_ json.Marshaler   = (*Array)(nil)
	_ json.Unmarshaler = (*Array)(nil)
)
